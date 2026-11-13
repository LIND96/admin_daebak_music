package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"admin_daebak_music/backend/internal/config"
)

func TestValidateChatRequest(t *testing.T) {
	ok := chatRelayRequest{
		DeviceID: "abc", AppsName: "bts",
		Messages: []chatIncomingMessage{
			{Role: "system", Content: "persona"},
			{Role: "user", Content: "hi"},
		},
	}
	if err := validateChatRequest(ok); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	for name, mutate := range map[string]func(*chatRelayRequest){
		"missing device":   func(r *chatRelayRequest) { r.DeviceID = " " },
		"missing app":      func(r *chatRelayRequest) { r.AppsName = "" },
		"no messages":      func(r *chatRelayRequest) { r.Messages = nil },
		"bad role":         func(r *chatRelayRequest) { r.Messages[1].Role = "tool" },
		"empty content":    func(r *chatRelayRequest) { r.Messages[1].Content = "  " },
		"oversize content": func(r *chatRelayRequest) { r.Messages[1].Content = strings.Repeat("x", maxChatCharsEach+1) },
	} {
		r := ok
		r.Messages = append([]chatIncomingMessage(nil), ok.Messages...)
		mutate(&r)
		if err := validateChatRequest(r); err == nil {
			t.Fatalf("%s: expected error", name)
		}
	}
	many := ok
	for i := 0; i < maxChatMessages; i++ {
		many.Messages = append(many.Messages, chatIncomingMessage{Role: "user", Content: "x"})
	}
	if err := validateChatRequest(many); err == nil {
		t.Fatal("too many messages: expected error")
	}
}

func TestToChatCompletionBody(t *testing.T) {
	body := toChatCompletionBody("m", []chatIncomingMessage{
		{Role: "system", Content: "be nice"},
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	})
	if body.Model != "m" {
		t.Fatal("model must pass through")
	}
	if len(body.Messages) != 3 || body.Messages[0].Role != "system" || body.Messages[2].Role != "assistant" {
		t.Fatalf("roles must pass through: %+v", body.Messages)
	}
	if body.MaxTokens <= 0 {
		t.Fatal("max tokens must be set")
	}
}

func TestCallProviderReturnsServedModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Errorf("auth = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"free-model-x","choices":[{"message":{"content":"hi"}}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`))
	}))
	defer srv.Close()
	h := &ChatHandler{}
	text, tIn, tOut, served, err := h.callProvider(context.Background(), srv.URL, "k", "requested-model",
		toChatCompletionBody("requested-model", []chatIncomingMessage{{Role: "user", Content: "hi"}}))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if text != "hi" || tIn != 5 || tOut != 2 {
		t.Fatalf("got %q %d %d", text, tIn, tOut)
	}
	if served != "free-model-x" {
		t.Fatalf("served = %q, want upstream model", served)
	}
}

func TestIsDirectGeminiURL(t *testing.T) {
	if !isDirectGeminiURL("https://generativelanguage.googleapis.com/v1beta/openai") {
		t.Fatal("Gemini endpoint must be blocked")
	}
	if isDirectGeminiURL("http://nine-router:20128/v1") {
		t.Fatal("9Router must not be blocked")
	}
}

func TestPickModelRoundRobin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Errorf("auth = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"b"},{"id":"a"},{"id":"b"},{"id":"  "},{"id":"c"}]}`))
	}))
	defer srv.Close()
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: srv.URL, ChatProviderAPIKey: "k"}}
	var got []string
	for i := 0; i < 4; i++ {
		m, err := h.pickModel(context.Background())
		if err != nil {
			t.Fatalf("pick %d err = %v", i, err)
		}
		got = append(got, m)
	}
	// Urutan dipertahankan, duplikat/kosong dibuang: [b a c], lalu wrap.
	want := []string{"b", "a", "c", "b"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pick %d = %q, want %q (all=%v)", i, got[i], want[i], got)
		}
	}
}

func TestPickModelUsesStaleCacheOnFetchFailure(t *testing.T) {
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: "http://127.0.0.1:1", ChatProviderAPIKey: "k"}}
	h.cachedModels = []string{"stale-model"}
	h.modelsFetchedAt = time.Now().Add(-time.Hour) // kadaluarsa -> fetch gagal -> pakai basi
	m, err := h.pickModel(context.Background())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if m != "stale-model" {
		t.Fatalf("m = %q, want stale cache", m)
	}
}

func TestPickModelEmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer srv.Close()
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: srv.URL, ChatProviderAPIKey: "k"}}
	if _, err := h.pickModel(context.Background()); err == nil {
		t.Fatal("empty model list must error")
	}
}
