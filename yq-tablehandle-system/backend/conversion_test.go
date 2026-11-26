package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConversionAPI(t *testing.T) {
	for _, tc := range []struct {
		name, reply string
		status      int
		want        string
	}{
		{"direct", `{"converted":{"id":"001","n":9007199254740993}}`, 200, "{\n  \"id\": \"001\",\n  \"n\": 9007199254740993\n}"},
		{"wrapped", `{"converted":[false,null]}`, 200, "[\n  false,\n  null\n]"},
		{"invalid", `{"message":"not JSON"}`, 502, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload struct {
					Messages []struct{ Content string }
					Model    string
					Format   struct{ Type string } `json:"response_format"`
				}
				if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-token" || json.NewDecoder(r.Body).Decode(&payload) != nil || len(payload.Messages) != 2 || payload.Messages[1].Content != "id=001" || payload.Model != "test-model" || payload.Format.Type != "json_object" || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
					t.Error("unexpected RPC request")
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "mock", "model": "test-model", "object": "chat.completion", "choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": tc.reply}, "finish_reason": "stop"}}})
			}))
			defer rpc.Close()
			t.Setenv("DEEPSEEK_BASE_URL", rpc.URL)
			t.Setenv("DEEPSEEK_MODEL", "test-model")
			t.Setenv("DEEPSEEK_API_KEY", "test-token")
			response := httptest.NewRecorder()
			router(t.TempDir()).ServeHTTP(response, httptest.NewRequest("POST", "/api/json-conversions", strings.NewReader(`{"content":"id=001"}`)))
			if response.Code != tc.status || tc.want != "" && response.Body.String() != tc.want {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body)
			}
		})
	}
}
