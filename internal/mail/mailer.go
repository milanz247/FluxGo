// Package mail sends transactional authentication messages.
package mail

import "log"

// Mailer delivers verification and password reset links.
type Mailer interface {
	SendVerification(email, link string) error
	SendPasswordReset(email, link string) error
}

// LogMailer writes development email links to the application log.
type LogMailer struct{}

func (LogMailer) SendVerification(email, link string) error {
	log.Printf("verification email to=%s link=%s", email, link)
	return nil
}

func (LogMailer) SendPasswordReset(email, link string) error {
	log.Printf("password reset email to=%s link=%s", email, link)
	return nil
}
