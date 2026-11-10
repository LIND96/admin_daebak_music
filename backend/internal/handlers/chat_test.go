package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestSecondsUntilMidnight(t *testing.T) {
	loc := jakartaLoc
	// 23:00 WIB -> 3600s to midnight.
	n := time.Date(2026, 11, 8, 23, 0, 0, 0, loc)
	if got := secondsUntilMidnight(n); got != 3600 {
		t.Fatalf("got %d, want 3600", got)
	}
	if got := secondsUntilMidnight(time.Date(2026, 11, 8, 0, 0, 0, 0, loc)); got != 86400 {
		t.Fatalf("got %d, want 86400", got)
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
