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

func TestToGeminiBody(t *testing.T) {
	body := toGeminiBody([]chatIncomingMessage{
		{Role: "system", Content: "be nice"},
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	})
	if body.SystemInstruction == nil || body.SystemInstruction.Parts[0].Text != "be nice" {
		t.Fatal("system message must become system_instruction")
	}
	if len(body.Contents) != 2 || body.Contents[0].Role != "user" || body.Contents[1].Role != "model" {
		t.Fatalf("role mapping wrong: %+v", body.Contents)
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
