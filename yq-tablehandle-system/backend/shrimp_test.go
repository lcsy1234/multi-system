package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShrimpConversion(t *testing.T) {
	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]string
		if r.Method != "POST" || r.URL.Path != "/rpc" || r.Header.Get("Authorization") != "Bearer example" || json.NewDecoder(r.Body).Decode(&in) != nil || !strings.Contains(in["content"], "id=0017") {
			t.Error("unexpected shrimp request")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"content": "```json\n{\"converted\":{\"id\":\"0017\"}}\n```"})
	}))
	defer rpc.Close()
	t.Setenv("XIAOXIAMI_TOKEN", "example")
	t.Setenv("XIAOXIAMI_RPC_URL", rpc.URL+"/rpc")
	res := httptest.NewRecorder()
	router(t.TempDir()).ServeHTTP(res, httptest.NewRequest("POST", "/api/shrimp/json-conversions", strings.NewReader(`{"content":"id=0017"}`)))
	if res.Code != 200 || res.Body.String() != "{\n  \"id\": \"0017\"\n}" {
		t.Fatal(res.Code, res.Body.String())
	}
}
