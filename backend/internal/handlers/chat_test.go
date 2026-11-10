package handlers

import (
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
