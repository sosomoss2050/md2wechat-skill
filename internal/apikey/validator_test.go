package apikey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidatorUsesHEADAuthValidate(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := NewValidator(server.URL, "secret-key").Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if gotMethod != http.MethodHead || gotPath != "/api/auth/validate" || gotAuth != "Bearer secret-key" {
		t.Fatalf("request = %s %s auth=%q", gotMethod, gotPath, gotAuth)
	}
}

func TestValidatorErrors(t *testing.T) {
	if err := NewValidator("https://example.com", "").Validate(context.Background()); !IsRequired(err) {
		t.Fatalf("missing key error = %v", err)
	}

	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer invalidServer.Close()
	if err := NewValidator(invalidServer.URL, "bad-key").Validate(context.Background()); !IsInvalid(err) {
		t.Fatalf("invalid key error = %v", err)
	}

	verifyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	verifyServer.Close()
	if err := NewValidator(verifyServer.URL, "key").Validate(context.Background()); !IsVerifyFailed(err) {
		t.Fatalf("verify error = %T %v", err, err)
	}
}

func TestNewValidatorWithTimeout(t *testing.T) {
	validator := NewValidatorWithTimeout("https://example.com", "key", 3*time.Second)
	if validator.client.Timeout != 3*time.Second {
		t.Fatalf("timeout = %s, want 3s", validator.client.Timeout)
	}

	defaultValidator := NewValidatorWithTimeout("https://example.com", "key", 0)
	if defaultValidator.client.Timeout != 10*time.Second {
		t.Fatalf("default timeout = %s, want 10s", defaultValidator.client.Timeout)
	}
}
