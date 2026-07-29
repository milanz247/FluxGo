package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"fluxgo/app/models"
	Route "fluxgo/internal/route"
	"fluxgo/internal/validation"
	"gorm.io/gorm"
)

type productInput struct {
	Name  string  `validate:"required,max=150" label:"Name"`
	Price float64 `validate:"min=0" label:"Price"`
	Qty   int     `validate:"min=0" label:"Quantity"`
}

// ProductHandler manages CRUD operations for products.
type ProductHandler struct {
	database *gorm.DB
}

// NewProductHandler creates a ProductHandler backed by database.
func NewProductHandler(database *gorm.DB) *ProductHandler {
	return &ProductHandler{database: database}
}

// Index lists products, optionally filtered by the "q" search query against name.
func (handler *ProductHandler) Index(c *Route.Context) error {
	search := strings.TrimSpace(c.Query("q"))

	query := handler.database.WithContext(c.Request.Context()).Model(&models.Product{}).Order("name")
	if search != "" {
		query = query.Where("name LIKE ? ESCAPE '"+likeEscapeChar+"'", "%"+escapeLike(search)+"%")
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		return Route.InternalServerError("could not load products", err)
	}

	return c.Render("products/index", Route.Data{
		"Title":    "Products",
		"Products": products,
		"Search":   search,
	})
}

// ShowCreate renders the add-product form.
func (handler *ProductHandler) ShowCreate(c *Route.Context) error {
	return c.Render("products/create", Route.Data{"Title": "Add product"})
}

// Store validates and creates a new product.
func (handler *ProductHandler) Store(c *Route.Context) error {
	name := strings.TrimSpace(c.Form("name"))
	priceRaw := c.Form("price")
	qtyRaw := c.Form("qty")
	data := Route.Data{
		"Title": "Add product", "OldName": name, "OldPrice": priceRaw, "OldQty": qtyRaw,
	}

	price, qty, ok := parseProductForm(priceRaw, qtyRaw)
	if !ok {
		return renderFormError(c, "products/create", data, "Price and quantity must be valid numbers.")
	}

	input := productInput{Name: name, Price: price, Qty: qty}
	if err := validation.Validate(input); err != nil {
		var errs validation.Errors
		errors.As(err, &errs)
		return renderFormError(c, "products/create", data,
			firstError(errs, "Name", "Price", "Quantity"))
	}

	product := models.Product{Name: name, Price: price, Qty: qty}
	if err := handler.database.WithContext(c.Request.Context()).Create(&product).Error; err != nil {
		return Route.InternalServerError("could not create product", err)
	}
	return c.Redirect("/products", http.StatusSeeOther)
}

// ShowEdit renders the edit-product form.
func (handler *ProductHandler) ShowEdit(c *Route.Context) error {
	product, err := handler.findProduct(c)
	if err != nil {
		return err
	}
	return c.Render("products/edit", Route.Data{
		"Title":    "Edit product",
		"Product":  product,
		"OldName":  product.Name,
		"OldPrice": formatFloat(product.Price),
		"OldQty":   strconv.Itoa(product.Qty),
	})
}

// Update validates and saves changes to an existing product.
func (handler *ProductHandler) Update(c *Route.Context) error {
	product, err := handler.findProduct(c)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(c.Form("name"))
	priceRaw := c.Form("price")
	qtyRaw := c.Form("qty")
	data := Route.Data{
		"Title": "Edit product", "Product": product,
		"OldName": name, "OldPrice": priceRaw, "OldQty": qtyRaw,
	}

	price, qty, ok := parseProductForm(priceRaw, qtyRaw)
	if !ok {
		return renderFormError(c, "products/edit", data, "Price and quantity must be valid numbers.")
	}

	input := productInput{Name: name, Price: price, Qty: qty}
	if err := validation.Validate(input); err != nil {
		var errs validation.Errors
		errors.As(err, &errs)
		return renderFormError(c, "products/edit", data,
			firstError(errs, "Name", "Price", "Quantity"))
	}

	updates := map[string]any{"name": name, "price": price, "qty": qty}
	if err := handler.database.WithContext(c.Request.Context()).
		Model(&product).Updates(updates).Error; err != nil {
		return Route.InternalServerError("could not update product", err)
	}
	return c.Redirect("/products", http.StatusSeeOther)
}

// Delete removes a product.
func (handler *ProductHandler) Delete(c *Route.Context) error {
	result := handler.database.WithContext(c.Request.Context()).
		Delete(&models.Product{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		return Route.InternalServerError("could not delete product", result.Error)
	}
	if result.RowsAffected == 0 {
		return Route.NotFoundError("that product does not exist")
	}
	return c.Redirect("/products", http.StatusSeeOther)
}

func (handler *ProductHandler) findProduct(c *Route.Context) (models.Product, error) {
	var product models.Product
	err := handler.database.WithContext(c.Request.Context()).
		First(&product, "id = ?", c.Param("id")).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return product, Route.NotFoundError("that product does not exist")
	}
	if err != nil {
		return product, Route.InternalServerError("could not load product", err)
	}
	return product, nil
}

func parseProductForm(priceRaw, qtyRaw string) (price float64, qty int, ok bool) {
	price, err := strconv.ParseFloat(strings.TrimSpace(priceRaw), 64)
	if err != nil {
		return 0, 0, false
	}
	qty, err = strconv.Atoi(strings.TrimSpace(qtyRaw))
	if err != nil {
		return 0, 0, false
	}
	return price, qty, true
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// likeEscapeChar marks LIKE wildcards as literal in escapeLike. It avoids
// backslash on purpose: MySQL unescapes "\\" inside a string literal to a
// single backslash while SQLite does not, so a backslash escape character
// behaves differently per database. "!" has no special meaning in either
// dialect's string literals, so ESCAPE '!' is portable.
const likeEscapeChar = "!"

// escapeLike escapes SQL LIKE wildcards in user input so search terms are
// matched literally instead of being interpreted as patterns.
func escapeLike(value string) string {
	replacer := strings.NewReplacer(
		likeEscapeChar, likeEscapeChar+likeEscapeChar,
		"%", likeEscapeChar+"%",
		"_", likeEscapeChar+"_",
	)
	return replacer.Replace(value)
}
