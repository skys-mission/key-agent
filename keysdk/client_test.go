package keysdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetStringOr(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		mockStatus int
		mockBody   interface{}
		want       string
	}{
		{
			name:       "key exists returns value",
			key:        "app/config",
			defaultVal: "default",
			mockStatus: http.StatusOK,
			mockBody:   KVEntry{Key: "app/config", Value: "actual-value"},
			want:       "actual-value",
		},
		{
			name:       "key not found returns default",
			key:        "missing/key",
			defaultVal: "default-value",
			mockStatus: http.StatusNotFound,
			mockBody:   errorResponse{Error: struct{ Code string `json:"code"`; Message string `json:"message"` }{Code: "NOT_FOUND", Message: "key not found"}},
			want:       "default-value",
		},
		{
			name:       "server error returns default",
			key:        "any/key",
			defaultVal: "fallback",
			mockStatus: http.StatusInternalServerError,
			mockBody:   errorResponse{Error: struct{ Code string `json:"code"`; Message string `json:"message"` }{Code: "INTERNAL", Message: "server error"}},
			want:       "fallback",
		},
		{
			name:       "empty default allowed",
			key:        "missing",
			defaultVal: "",
			mockStatus: http.StatusNotFound,
			mockBody:   errorResponse{Error: struct{ Code string `json:"code"`; Message string `json:"message"` }{Code: "NOT_FOUND", Message: "not found"}},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client := NewClient(&Config{
				BaseURL: server.URL,
				Token:   "test-token",
			})

			got := client.GetStringOr(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetStringOr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClient_GetSecretStringOr(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		mockStatus int
		mockBody   interface{}
		want       string
	}{
		{
			name:       "secret exists returns value",
			key:        "db/password",
			defaultVal: "default",
			mockStatus: http.StatusOK,
			mockBody:   SecretEntry{KVEntry: KVEntry{Key: "db/password", Value: "secret123"}, Type: SecretTypePassword},
			want:       "secret123",
		},
		{
			name:       "secret not found returns default",
			key:        "missing/secret",
			defaultVal: "fallback-secret",
			mockStatus: http.StatusNotFound,
			mockBody:   errorResponse{Error: struct{ Code string `json:"code"`; Message string `json:"message"` }{Code: "NOT_FOUND", Message: "secret not found"}},
			want:       "fallback-secret",
		},
		{
			name:       "unauthorized returns default",
			key:        "any/secret",
			defaultVal: "",
			mockStatus: http.StatusUnauthorized,
			mockBody:   errorResponse{Error: struct{ Code string `json:"code"`; Message string `json:"message"` }{Code: "UNAUTHORIZED", Message: "invalid token"}},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client := NewClient(&Config{
				BaseURL: server.URL,
				Token:   "test-token",
			})

			got := client.GetSecretStringOr(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetSecretStringOr() = %q, want %q", got, tt.want)
			}
		})
	}
}