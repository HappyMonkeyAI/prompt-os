package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIChatOmitsEmptyAuthorizationHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	got, err := openAIChat(context.Background(), server.URL, "", "local-model", "", "prompt")
	if err != nil {
		t.Fatalf("openAIChat returned error: %v", err)
	}
	if got != "ok" {
		t.Fatalf("openAIChat returned %q, want ok", got)
	}
	if gotAuth != "" {
		t.Fatalf("Authorization header = %q, want empty", gotAuth)
	}
}

func TestOpenAIChatSetsAuthorizationHeaderWhenKeyProvided(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer server.Close()

	_, err := openAIChat(context.Background(), server.URL, "sk-test", "gpt-test", "", "prompt")
	if err != nil {
		t.Fatalf("openAIChat returned error: %v", err)
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("Authorization header = %q, want Bearer sk-test", gotAuth)
	}
}
