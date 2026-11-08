package handlers

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"admin_daebak_music/backend/internal/config"
	"admin_daebak_music/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Chat relay: mobile POSTs here instead of calling a free AI endpoint
// directly, so per-IP lottery limits can't starve users. The server holds
// the provider key (never shipped in the APK) and enforces a fair daily
// quota per device_id. Contract mirrors the mobile ChatResult shape:
// success -> {"reply"}, quota -> 429 + retry_after_seconds, provider
// failure -> 502. Mobile falls back to direct free access on 502/network
// errors, but treats 429 as final (quota is quota).

type ChatHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

type chatIncomingMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRelayRequest struct {
	DeviceID string                `json:"device_id"`
	AppsName string                `json:"apps_name"`
	IdolName string                `json:"idol_name"`
	Messages []chatIncomingMessage `json:"messages"`
}

const (
	maxChatMessages   = 13
	maxChatCharsTotal = 6000
	maxChatCharsEach  = 4000
	maxReplyTokens    = 300
)

var validChatRoles = map[string]bool{"system": true, "user": true, "assistant": true}

// validateChatRequest is pure (unit-tested): shape and size limits only.
func validateChatRequest(req chatRelayRequest) error {
	if strings.TrimSpace(req.DeviceID) == "" || strings.TrimSpace(req.AppsName) == "" {
		return fmt.Errorf("device_id and apps_name are required")
	}
	if len(req.Messages) == 0 || len(req.Messages) > maxChatMessages {
		return fmt.Errorf("messages must contain 1-%d items", maxChatMessages)
	}
	total := 0
	for i, m := range req.Messages {
		if !validChatRoles[m.Role] {
			return fmt.Errorf("messages[%d]: invalid role %q", i, m.Role)
		}
		text := strings.TrimSpace(m.Content)
		if text == "" {
			return fmt.Errorf("messages[%d]: empty content", i)
		}
		if len(text) > maxChatCharsEach {
			return fmt.Errorf("messages[%d]: content too long", i)
		}
		total += len(text)
	}
	if total > maxChatCharsTotal {
		return fmt.Errorf("messages too long in total")
	}
	return nil
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  map[string]any  `json:"generationConfig,omitempty"`
}

// toGeminiBody maps OpenAI-style roles onto Gemini roles. First system
// message becomes system_instruction; the rest map user->user,
// assistant->model. Pure (unit-tested).
func toGeminiBody(msgs []chatIncomingMessage) geminiRequest {
	body := geminiRequest{Contents: []geminiContent{}}
	for _, m := range msgs {
		text := strings.TrimSpace(m.Content)
		if m.Role == "system" && body.SystemInstruction == nil {
			body.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: text}}}
			continue
		}
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		body.Contents = append(body.Contents, geminiContent{Role: role, Parts: []geminiPart{{Text: text}}})
	}
	body.GenerationConfig = map[string]any{"maxOutputTokens": maxReplyTokens, "temperature": 0.9}
	return body
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata *struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// secondsUntilMidnight returns seconds from now to next Jakarta midnight.
// Pure-ish (takes time, unit-tested) — drives the 429 retry_after_seconds.
func secondsUntilMidnight(now time.Time) int {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, jakartaLoc)
	secs := int(next.Sub(now).Seconds())
	if secs < 1 {
		return 1
	}
	return secs
}

func (h *ChatHandler) checkAppSecret(c *gin.Context) bool {
	want := strings.TrimSpace(h.Cfg.ChatAppSecret)
	if want == "" {
		return true // local dev convenience; set CHAT_APP_SECRET in prod
	}
	got := strings.TrimSpace(c.GetHeader("X-App-Secret"))
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	return true
}

func (h *ChatHandler) callGemini(ctx context.Context, body geminiRequest) (string, int, int, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/" +
		h.Cfg.GeminiModel + ":generateContent?key=" + h.Cfg.GeminiAPIKey
	payload, err := json.Marshal(body)
	if err != nil {
		return "", 0, 0, err
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", 0, 0, ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return "", 0, 0, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("provider status %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return "", 0, 0, fmt.Errorf("provider status %d", resp.StatusCode)
		}
		var out geminiResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			return "", 0, 0, err
		}
		if out.Error != nil {
			return "", 0, 0, fmt.Errorf("provider error: %s", out.Error.Message)
		}
		var sb strings.Builder
		for _, cand := range out.Candidates {
			for _, p := range cand.Content.Parts {
				sb.WriteString(p.Text)
			}
		}
		text := strings.TrimSpace(sb.String())
		if text == "" {
			return "", 0, 0, fmt.Errorf("empty provider reply")
		}
		tIn, tOut := 0, 0
		if out.UsageMetadata != nil {
			tIn, tOut = out.UsageMetadata.PromptTokenCount, out.UsageMetadata.CandidatesTokenCount
		}
		return text, tIn, tOut, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("provider unavailable")
	}
	return "", 0, 0, lastErr
}

// PostChat relays one chat message through the server-held provider key.
// Quota: CHAT_DAILY_LIMIT successes+failures per (device_id, apps_name, day).
func (h *ChatHandler) PostChat(c *gin.Context) {
	if h.Cfg.GeminiAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat relay disabled"})
		return
	}
	if !h.checkAppSecret(c) {
		return
	}
	var req chatRelayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateChatRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	appsName := strings.ToLower(strings.TrimSpace(req.AppsName))
	idolName := strings.TrimSpace(req.IdolName)
	if len(idolName) > 100 {
		idolName = idolName[:100]
	}
	now := jakartaNow()
	today := now.Format("2006-01-02")

	var used int64
	if err := h.DB.Model(&models.ChatMessage{}).
		Where("device_id = ? AND apps_name = ? AND usage_date = ?", deviceID, appsName, today).
		Count(&used).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "quota check failed"})
		return
	}
	if int(used) >= h.Cfg.ChatDailyLimit {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":               "daily quota exceeded",
			"retry_after_seconds": secondsUntilMidnight(now),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.Cfg.ChatTimeoutSec)*time.Second)
	defer cancel()
	reply, tIn, tOut, err := h.callGemini(ctx, toGeminiBody(req.Messages))

	row := models.ChatMessage{
		DeviceID:   deviceID,
		AppsName:   appsName,
		IdolName:   idolName,
		Provider:   "gemini:" + h.Cfg.GeminiModel,
		TokensIn:   tIn,
		TokensOut:  tOut,
		Ok:         err == nil,
		OccurredAt: time.Now().UTC(),
		UsageDate:  today,
	}
	_ = h.DB.Create(&row).Error // logging must not fail the reply

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "provider unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}
