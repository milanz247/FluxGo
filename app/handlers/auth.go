package handlers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"fluxgo/app/models"
	"fluxgo/internal/auth"
	AuthMail "fluxgo/internal/mail"
	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	minimumPasswordLength = 8
	maximumPasswordLength = 1024
)

type AuthHandler struct {
	database          *gorm.DB
	dummyPasswordHash string
	appURL            string
	mailer            AuthMail.Mailer
	limiter           *auth.LoginLimiter
	now               func() time.Time
}

func NewAuthHandler(
	database *gorm.DB,
	appURL string,
	mailer AuthMail.Mailer,
	limiter *auth.LoginLimiter,
) (*AuthHandler, error) {
	dummyHash, err := auth.HashPassword("not-a-real-user-password")
	if err != nil {
		return nil, err
	}
	return &AuthHandler{
		database: database, dummyPasswordHash: dummyHash, appURL: appURL,
		mailer: mailer, limiter: limiter, now: time.Now,
	}, nil
}

func (handler *AuthHandler) Home(c *Route.Context) error {
	if c.Session().Has("user_id") {
		return c.Redirect("/dashboard", http.StatusSeeOther)
	}
	return c.Redirect("/login", http.StatusSeeOther)
}

func (handler *AuthHandler) ShowRegister(c *Route.Context) error {
	return c.Render("register", Route.Data{"Title": "Create account"})
}

func (handler *AuthHandler) Register(c *Route.Context) error {
	name := strings.TrimSpace(c.Form("name"))
	email := normalizeEmail(c.Form("email"))
	password := c.Form("password")
	confirmation := c.Form("password_confirmation")
	data := Route.Data{"Title": "Create account", "OldName": name, "OldEmail": email}

	if name == "" || !validEmail(email) {
		return renderAuthError(c, "register", data, "Enter a valid name and email address.")
	}
	if err := validatePassword(password, confirmation); err != nil {
		return renderAuthError(c, "register", data, err.Error())
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user := models.User{Name: name, Email: email, PasswordHash: passwordHash}
	if err := handler.database.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return renderAuthError(c, "register", data, "An account already exists for that email address.")
		}
		return err
	}
	if err := handler.issueVerification(user); err != nil {
		return err
	}
	if err := handler.authenticate(c, user.ID); err != nil {
		return err
	}
	return c.Redirect("/dashboard", http.StatusSeeOther)
}

func (handler *AuthHandler) ShowLogin(c *Route.Context) error {
	message := ""
	switch c.Query("status") {
	case "password-reset":
		message = "Your password has been reset. You can now sign in."
	case "verified":
		message = "Your email address has been verified."
	case "invalid-verification":
		message = "That verification link is invalid or expired."
	}
	return c.Render("login", Route.Data{"Title": "Sign in", "Message": message})
}

func (handler *AuthHandler) Login(c *Route.Context) error {
	email := normalizeEmail(c.Form("email"))
	password := c.Form("password")
	data := Route.Data{"Title": "Sign in", "OldEmail": email}
	key := handler.limiter.Key(clientIP(c.Request), email)

	allowed, retryAfter, err := handler.limiter.Check(key, handler.now())
	if err != nil {
		return err
	}
	if !allowed {
		seconds := int(retryAfter.Round(time.Second).Seconds())
		if seconds < 1 {
			seconds = 1
		}
		c.SetHeader("Retry-After", strconv.Itoa(seconds))
		return c.RenderStatus(http.StatusTooManyRequests, "login", data.Set(
			"Error", "Too many login attempts. Try again later.",
		))
	}

	var user models.User
	err = handler.database.WithContext(c.Request.Context()).Where("email = ?", email).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	passwordHash := user.PasswordHash
	if err != nil {
		passwordHash = handler.dummyPasswordHash
	}
	if !auth.VerifyPassword(password, passwordHash) || err != nil {
		if err := handler.limiter.RecordFailure(key, handler.now()); err != nil {
			return err
		}
		return renderAuthError(c, "login", data, "The email address or password is incorrect.")
	}

	if err := handler.limiter.Clear(key); err != nil {
		return err
	}
	if err := handler.authenticate(c, user.ID); err != nil {
		return err
	}
	return c.Redirect("/dashboard", http.StatusSeeOther)
}

func (handler *AuthHandler) Dashboard(c *Route.Context) error {
	user, err := handler.currentUser(c)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = c.Session().Destroy()
			return c.Redirect("/login", http.StatusSeeOther)
		}
		return err
	}
	return c.Render("dashboard", Route.Data{
		"Title":    "Dashboard",
		"User":     user,
		"Verified": user.EmailVerifiedAt != nil,
		"Message":  dashboardMessage(c.Query("status")),
	})
}

func (handler *AuthHandler) Logout(c *Route.Context) error {
	if err := c.Session().Destroy(); err != nil {
		return err
	}
	return c.Redirect("/login", http.StatusSeeOther)
}

func (handler *AuthHandler) ShowForgotPassword(c *Route.Context) error {
	return c.Render("forgot-password", Route.Data{"Title": "Forgot password"})
}

func (handler *AuthHandler) ForgotPassword(c *Route.Context) error {
	email := normalizeEmail(c.Form("email"))
	var user models.User
	err := handler.database.WithContext(c.Request.Context()).Where("email = ?", email).First(&user).Error
	if err == nil {
		raw, hash, tokenErr := auth.NewToken()
		if tokenErr != nil {
			return tokenErr
		}
		if err := handler.database.Where("user_id = ? AND used_at IS NULL", user.ID).
			Delete(&models.PasswordReset{}).Error; err != nil {
			return err
		}
		reset := models.PasswordReset{
			UserID: user.ID, TokenHash: hash, ExpiresAt: handler.now().Add(time.Hour),
		}
		if err := handler.database.Create(&reset).Error; err != nil {
			return err
		}
		link := handler.appURL + "/reset-password?token=" + raw
		if err := handler.mailer.SendPasswordReset(user.Email, link); err != nil {
			return err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return c.Render("forgot-password", Route.Data{
		"Title":   "Forgot password",
		"Message": "If that email is registered, a password reset link has been sent.",
	})
}

func (handler *AuthHandler) ShowResetPassword(c *Route.Context) error {
	token := c.Query("token")
	if _, err := handler.validReset(token); err != nil {
		return c.RenderStatus(http.StatusUnprocessableEntity, "reset-password", Route.Data{
			"Title": "Reset password", "Error": "This reset link is invalid or expired.",
		})
	}
	return c.Render("reset-password", Route.Data{"Title": "Reset password", "Token": token})
}

func (handler *AuthHandler) ResetPassword(c *Route.Context) error {
	token := c.Form("token")
	data := Route.Data{"Title": "Reset password", "Token": token}
	password := c.Form("password")
	if err := validatePassword(password, c.Form("password_confirmation")); err != nil {
		return renderAuthError(c, "reset-password", data, err.Error())
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	now := handler.now()
	err = handler.database.Transaction(func(tx *gorm.DB) error {
		var reset models.PasswordReset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?",
				auth.HashToken(token), now).
			First(&reset).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", reset.UserID).
			Update("password_hash", hash).Error; err != nil {
			return err
		}
		if err := tx.Model(&reset).Update("used_at", now).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", reset.UserID).Delete(&session.DatabaseRecord{}).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return renderAuthError(c, "reset-password", data, "This reset link is invalid or expired.")
	}
	if err != nil {
		return err
	}
	_ = c.Session().Destroy()
	return c.Redirect("/login?status=password-reset", http.StatusSeeOther)
}

func (handler *AuthHandler) VerifyEmail(c *Route.Context) error {
	hash := auth.HashToken(c.Query("token"))
	now := handler.now()
	err := handler.database.Transaction(func(tx *gorm.DB) error {
		var verification models.EmailVerification
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, now).
			First(&verification).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", verification.UserID).
			Update("email_verified_at", now).Error; err != nil {
			return err
		}
		return tx.Model(&verification).Update("used_at", now).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Redirect("/login?status=invalid-verification", http.StatusSeeOther)
	}
	if err != nil {
		return err
	}
	if c.Session().Has("user_id") {
		return c.Redirect("/dashboard?status=verified", http.StatusSeeOther)
	}
	return c.Redirect("/login?status=verified", http.StatusSeeOther)
}

func (handler *AuthHandler) ResendVerification(c *Route.Context) error {
	user, err := handler.currentUser(c)
	if err != nil {
		return err
	}
	if user.EmailVerifiedAt == nil {
		if err := handler.issueVerification(user); err != nil {
			return err
		}
	}
	return c.Redirect("/dashboard?status=verification-sent", http.StatusSeeOther)
}

func (handler *AuthHandler) issueVerification(user models.User) error {
	raw, hash, err := auth.NewToken()
	if err != nil {
		return err
	}
	if err := handler.database.Where("user_id = ? AND used_at IS NULL", user.ID).
		Delete(&models.EmailVerification{}).Error; err != nil {
		return err
	}
	verification := models.EmailVerification{
		UserID: user.ID, TokenHash: hash, ExpiresAt: handler.now().Add(24 * time.Hour),
	}
	if err := handler.database.Create(&verification).Error; err != nil {
		return err
	}
	return handler.mailer.SendVerification(user.Email, handler.appURL+"/email/verify?token="+raw)
}

func (handler *AuthHandler) validReset(token string) (models.PasswordReset, error) {
	var reset models.PasswordReset
	if token == "" {
		return reset, gorm.ErrRecordNotFound
	}
	err := handler.database.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?",
		auth.HashToken(token), handler.now()).First(&reset).Error
	return reset, err
}

func (handler *AuthHandler) currentUser(c *Route.Context) (models.User, error) {
	userID, _ := c.Session().Get("user_id")
	var user models.User
	err := handler.database.WithContext(c.Request.Context()).First(&user, userID).Error
	return user, err
}

func (handler *AuthHandler) authenticate(c *Route.Context, userID uint) error {
	if err := c.Session().Regenerate(); err != nil {
		return err
	}
	return c.Session().Set("user_id", userID)
}

func renderAuthError(c *Route.Context, view string, data Route.Data, message string) error {
	data["Error"] = message
	return c.RenderStatus(http.StatusUnprocessableEntity, view, data)
}

func validatePassword(password, confirmation string) error {
	if len(password) < minimumPasswordLength || len(password) > maximumPasswordLength {
		return fmt.Errorf("password must be between 8 and 1024 characters")
	}
	if password != confirmation {
		return fmt.Errorf("password confirmation does not match")
	}
	return nil
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}

func dashboardMessage(status string) string {
	switch status {
	case "verified":
		return "Your email address has been verified."
	case "verification-sent":
		return "A new verification link has been sent."
	default:
		return ""
	}
}
