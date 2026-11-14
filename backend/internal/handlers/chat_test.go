package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

func TestChatWithFailoverTriesNextModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"bad-model"},{"id":"good-model"}]}`))
		case "/chat/completions":
			var body struct {
				Model string `json:"model"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Model == "bad-model" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"model":"good-model","choices":[{"message":{"content":"hi"}}]}`))
		default:
			t.Errorf("path = %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: srv.URL, ChatProviderAPIKey: "k"}}
	reply, _, _, requested, served, err := h.chatWithFailover(context.Background(), srv.URL, "k",
		[]chatIncomingMessage{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if reply != "hi" || requested != "good-model" || served != "good-model" {
		t.Fatalf("got reply=%q requested=%q served=%q", reply, requested, served)
	}
}

func TestDecodeErrorContextOffset(t *testing.T) {
	raw := []byte(`{"ok":true}data: broken-tail`)
	var v map[string]any
	err := json.Unmarshal(raw, &v)
	if err == nil {
		t.Fatal("expected syntax error")
	}
	ctx := decodeErrorContext(raw, err)
	if !strings.Contains(ctx, "offset 12") || !strings.Contains(ctx, "data: broken-tail") {
		t.Fatalf("ctx = %q", ctx)
	}
}

func TestDecodeChatResponseIgnoresTrailingSSE(t *testing.T) {
	// Bentuk persis dari 9Router: JSON lengkap + frame SSE di ekor.
	raw := []byte("{\"model\":\"m\",\"choices\":[{\"message\":{\"content\":\"hi\"}}]," +
		"\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2}}data: [DONE]\n\n")
	out, err := decodeChatResponse(raw)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(out.Choices) != 1 || replyTextOf(out.Choices[0].Message.Content) != "hi" {
		t.Fatalf("out = %+v", out)
	}
	if out.Usage == nil || out.Usage.PromptTokens != 1 || out.Usage.CompletionTokens != 2 {
		t.Fatalf("usage = %+v", out.Usage)
	}
}

func TestDecodeChatResponseRejectsGarbage(t *testing.T) {
	if _, err := decodeChatResponse([]byte("data: broken-tail")); err == nil {
		t.Fatal("non-JSON must error")
	}
	if _, err := decodeChatResponse([]byte(`{"model":`)); err == nil {
		t.Fatal("truncated JSON must error")
	}
}

func TestChatWithFailoverAllFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/models" {
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"m1"},{"id":"m2"}]}`))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: srv.URL, ChatProviderAPIKey: "k"}}
	_, _, _, requested, _, err := h.chatWithFailover(context.Background(), srv.URL, "k",
		[]chatIncomingMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("all models failing must error")
	}
	if requested == "" {
		t.Fatal("requested must record last tried model")
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

func TestChatWithFailoverSkipsCooldownModel(t *testing.T) {
	var mu sync.Mutex
	hits := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/models" {
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"bad-model"},{"id":"good-model"}]}`))
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		hits[body.Model]++
		mu.Unlock()
		if body.Model == "bad-model" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"model":"good-model","choices":[{"message":{"content":"hi"}}]}`))
	}))
	defer srv.Close()
	h := &ChatHandler{Cfg: &config.Config{ChatProviderBaseURL: srv.URL, ChatProviderAPIKey: "k"}}
	msgs := []chatIncomingMessage{{Role: "user", Content: "hi"}}
	if _, _, _, _, _, err := h.chatWithFailover(context.Background(), srv.URL, "k", msgs); err != nil {
		t.Fatalf("first err = %v", err)
	}
	// Request kedua: bad-model sedang cooldown -> tidak boleh dicoba lagi.
	if _, _, _, requested, _, err := h.chatWithFailover(context.Background(), srv.URL, "k", msgs); err != nil {
		t.Fatalf("second err = %v", err)
	} else if requested != "good-model" {
		t.Fatalf("requested = %q, want good-model (bad in cooldown)", requested)
	}
	mu.Lock()
	defer mu.Unlock()
	if hits["bad-model"] != 1 {
		t.Fatalf("bad-model hits = %d, want 1 (skipped while in cooldown)", hits["bad-model"])
	}
}
