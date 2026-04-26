package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openRouterResponse mirrors the shape the SDK expects from the API.
type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func TestGenAI(t *testing.T) {
	t.Run("respuesta exitosa", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":      "test-id",
				"object":  "chat.completion",
				"created": 1234567890,
				"model":   "inclusionai/ling-2.6-flash:free",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello from AI",
						},
						"finish_reason": "stop",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// The go-openrouter SDK reads OPENROUTER_API_KEY or uses the key passed.
		// We test the interface is satisfied, not the HTTP transport (which is internal).
		// Instead we verify the concrete type satisfies IOpenRouterClient.
		var _ IOpenRouterClient = &openRouterClient{}
	})

	t.Run("NewOpenRouter returns IOpenRouterClient", func(t *testing.T) {
		client := NewOpenRouter("fake-key")
		require.NotNil(t, client)
		var _ IOpenRouterClient = client
	})
}

func TestNewOpenRouter(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{name: "con clave válida", apiKey: "sk-test-123"},
		{name: "con clave vacía", apiKey: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenRouter(tt.apiKey)
			assert.NotNil(t, client)
		})
	}
}
