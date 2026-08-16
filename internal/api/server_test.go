package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"weddinglive/internal/api"
	"weddinglive/internal/domain"
	"weddinglive/internal/service"
	"weddinglive/internal/store"
)

func TestHTTPWorkflowAndConfigIsolation(t *testing.T) {
	handler := api.New(service.New(store.NewMemory(domain.State{}), "admin-token")).Handler()

	configRequest := httptest.NewRequest(http.MethodGet, "/config.example.json", nil)
	configResponse := httptest.NewRecorder()
	handler.ServeHTTP(configResponse, configRequest)
	if configResponse.Code != http.StatusNotFound {
		t.Fatalf("config response status = %d", configResponse.Code)
	}

	accountRequest := httptest.NewRequest(http.MethodPost, "/api/admin/accounts", bytes.NewBufferString(`{"name":"API 摄影师"}`))
	accountRequest.Header.Set("X-Admin-Token", "admin-token")
	accountResponse := httptest.NewRecorder()
	handler.ServeHTTP(accountResponse, accountRequest)
	if accountResponse.Code != http.StatusCreated {
		t.Fatalf("account response status = %d, body = %s", accountResponse.Code, accountResponse.Body.String())
	}

	roomRequest := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewBufferString(`{"title":"API 婚礼"}`))
	roomRequest.Header.Set("X-Photographer-Token", "photo-token-001")
	roomResponse := httptest.NewRecorder()
	handler.ServeHTTP(roomResponse, roomRequest)
	if roomResponse.Code != http.StatusCreated {
		t.Fatalf("room response status = %d, body = %s", roomResponse.Code, roomResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !bytes.Contains(listResponse.Body.Bytes(), []byte("API 婚礼")) {
		t.Fatalf("list response status = %d, body = %s", listResponse.Code, listResponse.Body.String())
	}
}
