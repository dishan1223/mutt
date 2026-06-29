package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dishan1223/mutt/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

func newTestAppWithProjects() *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/auth/signup", SignUpHandler)
	app.Post("/api/v1/auth/login", LoginHandler)
	app.Get("/api/v1/auth/me", middleware.AuthRequired, MeHandler)

	app.Post("/api/v1/projects/", middleware.AuthRequired, CreateProjectHandler)
	app.Get("/api/v1/projects/", middleware.AuthRequired, ListProjectsHandler)
	app.Get("/api/v1/projects/:id", middleware.AuthRequired, GetProjectHandler)
	app.Patch("/api/v1/projects/:id", middleware.AuthRequired, UpdateProjectHandler)
	app.Delete("/api/v1/projects/:id", middleware.AuthRequired, DeleteProjectHandler)
	app.Post("/api/v1/projects/:id/rotate-key", middleware.AuthRequired, RotateAPIKeyHandler)
	return app
}

func loginProjectUser(t *testing.T, app *fiber.App, email string) string {
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

func createProject(t *testing.T, app *fiber.App, token, name string) map[string]interface{} {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func TestCreateProject_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "projuser", "proj@example.com", "password123", "+9999999999")
	token := loginProjectUser(t, app, "proj@example.com")

	result := createProject(t, app, token, "My Project")

	if result["name"] != "My Project" {
		t.Fatalf("expected name 'My Project', got %v", result["name"])
	}
	if result["api_key"] == nil || result["api_key"] == "" {
		t.Fatal("expected api_key in response")
	}
}

func TestCreateProject_ValidationFails(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "projuser2", "proj2@example.com", "password123", "+1010101010")
	token := loginProjectUser(t, app, "proj2@example.com")

	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateProject_NoAuth(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()

	body, _ := json.Marshal(map[string]string{"name": "Unauthorized"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestListProjects_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "listuser", "list@example.com", "password123", "+1111111111")
	token := loginProjectUser(t, app, "list@example.com")

	createProject(t, app, token, "Project A")
	createProject(t, app, token, "Project B")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/", nil)
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
	projects := result["projects"].([]interface{})
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestListProjects_Empty(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "emptyuser", "empty@example.com", "password123", "+2222222222")
	token := loginProjectUser(t, app, "empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	projects := result["projects"].([]interface{})
	if len(projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(projects))
	}
}

func TestGetProject_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "getuser", "get@example.com", "password123", "+3333333333")
	token := loginProjectUser(t, app, "get@example.com")

	result := createProject(t, app, token, "Get Me")
	projectID := result["id"].(float64)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", int(projectID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var getResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&getResult)
	project := getResult["project"].(map[string]interface{})
	if project["name"] != "Get Me" {
		t.Fatalf("expected name 'Get Me', got %v", project["name"])
	}
}

func TestGetProject_NotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "getuser2", "get2@example.com", "password123", "+4444444444")
	token := loginProjectUser(t, app, "get2@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "upduser", "upd@example.com", "password123", "+5555555555")
	token := loginProjectUser(t, app, "upd@example.com")

	result := createProject(t, app, token, "Old Name")
	projectID := result["id"].(float64)

	updateBody, _ := json.Marshal(map[string]string{"name": "New Name"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/projects/%d", int(projectID)), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var updResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updResult)
	project := updResult["project"].(map[string]interface{})
	if project["name"] != "New Name" {
		t.Fatalf("expected name 'New Name', got %v", project["name"])
	}
}

func TestDeleteProject_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "deluser", "del@example.com", "password123", "+6666666666")
	token := loginProjectUser(t, app, "del@example.com")

	result := createProject(t, app, token, "Delete Me")
	projectID := result["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", int(projectID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// verify deleted
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", int(projectID)), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getResp, _ := app.Test(getReq)
	if getResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "deluser2", "del2@example.com", "password123", "+7777777777")
	token := loginProjectUser(t, app, "del2@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRotateAPIKey_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "rotuser", "rot@example.com", "password123", "+8888888888")
	token := loginProjectUser(t, app, "rot@example.com")

	result := createProject(t, app, token, "Rotate Me")
	projectID := result["id"].(float64)
	oldKey := result["api_key"].(string)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/rotate-key", int(projectID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var rotResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&rotResult)
	newKey := rotResult["api_key"].(string)
	if newKey == oldKey {
		t.Fatal("expected new key to differ from old key")
	}
}

func TestProject_UserCannotAccessOtherUsersProject(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()

	seedUser(t, "userA", "a@example.com", "password123", "+1111111111")
	tokenA := loginProjectUser(t, app, "a@example.com")
	result := createProject(t, app, tokenA, "User A Project")
	projectID := result["id"].(float64)

	seedUser(t, "userB", "b@example.com", "password123", "+2222222222")
	tokenB := loginProjectUser(t, app, "b@example.com")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", int(projectID)), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 for cross-user access, got %d", resp.StatusCode)
	}
}

func TestProject_UserIsolation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()

	seedUser(t, "iso1", "iso1@example.com", "password123", "+1111111111")
	token1 := loginProjectUser(t, app, "iso1@example.com")
	createProject(t, app, token1, "P1")
	createProject(t, app, token1, "P2")

	seedUser(t, "iso2", "iso2@example.com", "password123", "+2222222222")
	token2 := loginProjectUser(t, app, "iso2@example.com")
	createProject(t, app, token2, "P3")

	// user1 sees 2
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/projects/", nil)
	req1.Header.Set("Authorization", "Bearer "+token1)
	resp1, _ := app.Test(req1)
	var r1 map[string]interface{}
	json.NewDecoder(resp1.Body).Decode(&r1)
	if len(r1["projects"].([]interface{})) != 2 {
		t.Fatal("user1 should see 2 projects")
	}

	// user2 sees 1
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/projects/", nil)
	req2.Header.Set("Authorization", "Bearer "+token2)
	resp2, _ := app.Test(req2)
	var r2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&r2)
	if len(r2["projects"].([]interface{})) != 1 {
		t.Fatal("user2 should see 1 project")
	}
}

func TestRotateAPIKey_OldKeyInvalid(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	app := newTestAppWithProjects()
	seedUser(t, "rotuser2", "rot2@example.com", "password123", "+9999999999")
	token := loginProjectUser(t, app, "rot2@example.com")

	result := createProject(t, app, token, "Rotate Verify")
	projectID := result["id"].(float64)
	oldKey := result["api_key"].(string)

	// rotate
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/rotate-key", int(projectID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var rotResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&rotResult)
	newKey := rotResult["api_key"].(string)
	if newKey == oldKey {
		t.Fatal("expected new key to differ from old key")
	}
}
