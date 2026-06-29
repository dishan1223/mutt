package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/middleware"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/gofiber/fiber/v3"
)

func newTestApp() *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/auth/signup", SignUpHandler)
	app.Post("/api/v1/auth/login", LoginHandler)
	app.Post("/api/v1/auth/logout", middleware.AuthRequired, LogoutHandler)
	app.Post("/api/v1/auth/refresh", RefreshTokenHandler)
	app.Get("/api/v1/auth/me", middleware.AuthRequired, MeHandler)
	return app
}

func loginAndGetToken(t *testing.T, app *fiber.App, email string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "access_token" {
			return cookie.Value
		}
	}
	t.Fatal("access_token cookie not found")
	return ""
}

func TestSignUp_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
		"phone":    "+1234567890",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["message"] != "Success" {
		t.Fatalf("expected success message, got %v", result["message"])
	}
}

func TestSignUp_DuplicateEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "existing", "dup@example.com", "password123", "+1111111111")

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"email":    "dup@example.com",
		"password": "password123",
		"phone":    "+2222222222",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestSignUp_ValidationFails(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	body, _ := json.Marshal(map[string]string{
		"username": "",
		"email":    "not-an-email",
		"password": "short",
		"phone":    "",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLogin_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "logintest", "login@example.com", "password123", "+3333333333")

	body, _ := json.Marshal(map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["message"] != "Success" {
		t.Fatalf("expected success message, got %v", result["message"])
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "logintest2", "login2@example.com", "password123", "+4444444444")

	body, _ := json.Marshal(map[string]string{
		"email":    "login2@example.com",
		"password": "wrongpassword",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	body, _ := json.Marshal(map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMe_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "meuser", "me@example.com", "password123", "+5555555555")
	token := loginAndGetToken(t, app, "me@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	user := result["user"].(map[string]interface{})
	if user["email"] != "me@example.com" {
		t.Fatalf("expected email me@example.com, got %v", user["email"])
	}
}

func TestMe_NoAuth(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMe_InvalidToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogout_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "logoutuser", "logout@example.com", "password123", "+6666666666")
	token := loginAndGetToken(t, app, "logout@example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["message"] != "Logged out successfully" {
		t.Fatalf("expected logout message, got %v", result["message"])
	}
}

func TestRefresh_NoToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestSignUp_Login_Me_Flow(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()

	// signup
	signupBody, _ := json.Marshal(map[string]string{
		"username": "flowuser",
		"email":    "flow@example.com",
		"password": "password123",
		"phone":    "+7777777777",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody))
	signupReq.Header.Set("Content-Type", "application/json")
	signupResp, _ := app.Test(signupReq)
	if signupResp.StatusCode != fiber.StatusOK {
		t.Fatalf("signup failed: %d", signupResp.StatusCode)
	}

	// login
	token := loginAndGetToken(t, app, "flow@example.com")

	// me
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("me failed: %d", resp.StatusCode)
	}

	var meResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&meResult)
	user := meResult["user"].(map[string]interface{})
	if user["email"] != "flow@example.com" {
		t.Fatalf("expected email flow@example.com, got %v", user["email"])
	}
}

func TestBlacklistedTokenRejected(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	seedUser(t, "bluser", "bl@example.com", "password123", "+8888888888")
	token := loginAndGetToken(t, app, "bl@example.com")

	claims, _ := service.ValidateAccessToken(token)
	service.BlacklistAccessToken(claims.TokenID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for blacklisted token, got %d", resp.StatusCode)
	}
}

func TestMe_UserDeleted(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestApp()
	user := seedUser(t, "delme", "delme@example.com", "password123", "+9999999999")
	token := loginAndGetToken(t, app, "delme@example.com")

	// delete user from DB directly
	config.DB.Delete(&user)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
