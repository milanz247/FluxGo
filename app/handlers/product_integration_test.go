package handlers_test

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"fluxgo/app/handlers"
	AppMiddleware "fluxgo/app/middleware"
)

var productEditLinkPattern = regexp.MustCompile(`/products/(\d+)/edit`)

func TestProductCRUDFlow(t *testing.T) {
	app := newAuthApplicationWithProducts(t)
	cookie := registerUser(t, app, "products@example.com", "strong-password")

	createPage := app.request(http.MethodGet, "/products/create", nil, cookie)
	if createPage.Code != http.StatusOK {
		t.Fatalf("unexpected create page response: %d", createPage.Code)
	}
	createToken := extractCSRF(t, createPage.Body.String())

	create := app.request(http.MethodPost, "/products", url.Values{
		"_token": {createToken},
		"name":   {"Widget"},
		"price":  {"9.99"},
		"qty":    {"10"},
	}, cookie)
	if create.Code != http.StatusSeeOther || create.Header().Get("Location") != "/products" {
		t.Fatalf("unexpected create response: %d %s", create.Code, create.Body.String())
	}

	index := app.request(http.MethodGet, "/products", nil, cookie)
	if index.Code != http.StatusOK || !strings.Contains(index.Body.String(), "Widget") {
		t.Fatalf("expected the product list to contain Widget: %d %s", index.Code, index.Body.String())
	}

	searchHit := app.request(http.MethodGet, "/products?q=Wid", nil, cookie)
	if !strings.Contains(searchHit.Body.String(), "Widget") {
		t.Fatalf("expected search for %q to match Widget", "Wid")
	}
	searchMiss := app.request(http.MethodGet, "/products?q=nomatch", nil, cookie)
	if strings.Contains(searchMiss.Body.String(), "Widget") {
		t.Fatal("expected an unrelated search term to exclude Widget")
	}

	match := productEditLinkPattern.FindStringSubmatch(index.Body.String())
	if len(match) != 2 {
		t.Fatalf("could not find a product edit link in %s", index.Body.String())
	}
	id := match[1]

	editPage := app.request(http.MethodGet, "/products/"+id+"/edit", nil, cookie)
	if editPage.Code != http.StatusOK || !strings.Contains(editPage.Body.String(), "Widget") {
		t.Fatalf("unexpected edit page response: %d %s", editPage.Code, editPage.Body.String())
	}
	editToken := extractCSRF(t, editPage.Body.String())

	update := app.request(http.MethodPost, "/products/"+id, url.Values{
		"_token": {editToken},
		"name":   {"Widget Pro"},
		"price":  {"19.99"},
		"qty":    {"5"},
	}, cookie)
	if update.Code != http.StatusSeeOther || update.Header().Get("Location") != "/products" {
		t.Fatalf("unexpected update response: %d %s", update.Code, update.Body.String())
	}

	afterUpdate := app.request(http.MethodGet, "/products", nil, cookie)
	if !strings.Contains(afterUpdate.Body.String(), "Widget Pro") {
		t.Fatal("expected the product list to reflect the renamed product")
	}

	deleteToken := extractCSRF(t, afterUpdate.Body.String())
	deleteResponse := app.request(http.MethodPost, "/products/"+id+"/delete", url.Values{
		"_token": {deleteToken},
	}, cookie)
	if deleteResponse.Code != http.StatusSeeOther || deleteResponse.Header().Get("Location") != "/products" {
		t.Fatalf("unexpected delete response: %d %s", deleteResponse.Code, deleteResponse.Body.String())
	}

	afterDelete := app.request(http.MethodGet, "/products", nil, cookie)
	if strings.Contains(afterDelete.Body.String(), "Widget Pro") {
		t.Fatal("expected the product to be gone after delete")
	}
}

func TestProductRoutesRequireAuthentication(t *testing.T) {
	app := newAuthApplicationWithProducts(t)

	response := app.request(http.MethodGet, "/products", nil, nil)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/login" {
		t.Fatalf("expected an anonymous request to redirect to /login, got %d %s",
			response.Code, response.Header().Get("Location"))
	}
}

func TestProductValidationRejectsInvalidInput(t *testing.T) {
	app := newAuthApplicationWithProducts(t)
	cookie := registerUser(t, app, "invalid-product@example.com", "strong-password")

	createPage := app.request(http.MethodGet, "/products/create", nil, cookie)
	token := extractCSRF(t, createPage.Body.String())

	response := app.request(http.MethodPost, "/products", url.Values{
		"_token": {token},
		"name":   {""},
		"price":  {"-5"},
		"qty":    {"1"},
	}, cookie)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected validation to reject an empty name, got %d %s", response.Code, response.Body.String())
	}
}

// newAuthApplicationWithProducts extends the shared auth test application
// with the product CRUD routes registered under auth middleware, mirroring
// how route/web.go wires them in the real application.
func newAuthApplicationWithProducts(t *testing.T) *authApplication {
	t.Helper()
	app := newAuthApplication(t)

	productHandler := handlers.NewProductHandler(app.database)
	products := app.engine.Group("/products").Use(AppMiddleware.Auth)
	products.Get("", productHandler.Index)
	products.Get("/create", productHandler.ShowCreate)
	products.Post("", productHandler.Store)
	products.Get("/{id}/edit", productHandler.ShowEdit)
	products.Post("/{id}", productHandler.Update)
	products.Post("/{id}/delete", productHandler.Delete)

	return app
}
