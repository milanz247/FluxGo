package validation_test

import (
	"errors"
	"testing"

	"fluxgo/internal/validation"
)

type registerInput struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=20"`
	Confirm  string `validate:"required,eqfield=Password" label:"Password confirmation"`
}

func TestValidatePassesForValidInput(t *testing.T) {
	input := registerInput{
		Name: "Milan", Email: "milan@example.com",
		Password: "supersecret", Confirm: "supersecret",
	}
	if err := validation.Validate(input); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateReportsRequiredFields(t *testing.T) {
	err := validation.Validate(registerInput{})

	var errs validation.Errors
	if !errors.As(err, &errs) {
		t.Fatalf("expected validation.Errors, got %T", err)
	}
	for _, field := range []string{"Name", "Email", "Password", "Password confirmation"} {
		if !errs.Has(field) {
			t.Fatalf("expected an error for %q, got %v", field, errs)
		}
	}
}

func TestValidateEmailFormat(t *testing.T) {
	input := registerInput{
		Name: "Milan", Email: "not-an-email",
		Password: "supersecret", Confirm: "supersecret",
	}
	var errs validation.Errors
	if !errors.As(validation.Validate(input), &errs) {
		t.Fatal("expected a validation error for an invalid email")
	}
	if got := errs.First("Email"); got != "must be a valid email address" {
		t.Fatalf("unexpected message %q", got)
	}
}

func TestValidateMinMax(t *testing.T) {
	input := registerInput{
		Name: "Milan", Email: "milan@example.com",
		Password: "short", Confirm: "short",
	}
	var errs validation.Errors
	if !errors.As(validation.Validate(input), &errs) {
		t.Fatal("expected a validation error for a short password")
	}
	if got := errs.First("Password"); got != "must be at least 8 characters" {
		t.Fatalf("unexpected message %q", got)
	}
}

func TestValidateEqField(t *testing.T) {
	input := registerInput{
		Name: "Milan", Email: "milan@example.com",
		Password: "supersecret", Confirm: "different",
	}
	var errs validation.Errors
	if !errors.As(validation.Validate(input), &errs) {
		t.Fatal("expected a validation error for a mismatched confirmation")
	}
	if got := errs.First("Password confirmation"); got != "must match Password" {
		t.Fatalf("unexpected message %q", got)
	}
}

func TestValidateOneofAndNumeric(t *testing.T) {
	type input struct {
		Role string `validate:"required,oneof=admin|member"`
		PIN  string `validate:"required,numeric,len=4"`
	}

	var errs validation.Errors
	err := validation.Validate(input{Role: "owner", PIN: "12a4"})
	if !errors.As(err, &errs) {
		t.Fatal("expected validation errors")
	}
	if got := errs.First("Role"); got != "must be one of admin, member" {
		t.Fatalf("unexpected message %q", got)
	}
	if got := errs.First("PIN"); got != "must contain only digits" {
		t.Fatalf("unexpected message %q", got)
	}

	if err := validation.Validate(input{Role: "admin", PIN: "1234"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidatePointerAndNonStruct(t *testing.T) {
	input := &registerInput{
		Name: "Milan", Email: "milan@example.com",
		Password: "supersecret", Confirm: "supersecret",
	}
	if err := validation.Validate(input); err != nil {
		t.Fatalf("expected no error for a valid pointer, got %v", err)
	}

	if err := validation.Validate("not a struct"); err == nil {
		t.Fatal("expected an error for a non-struct value")
	}

	var nilInput *registerInput
	if err := validation.Validate(nilInput); err == nil {
		t.Fatal("expected an error for a nil pointer")
	}
}
