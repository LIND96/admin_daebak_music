package handlers

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []chatCompletionMessage `json:"messages"`
	Temperature float64                 `json:"temperature,omitempty"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
}

// toChatCompletionBody maps validated messages onto the OpenAI-compatible
// shape served by 9Router (/v1/chat/completions) and by Gemini's own
// OpenAI endpoint. Roles pass through (system/user/assistant). Pure
// (unit-tested).
func toChatCompletionBody(model string, msgs []chatIncomingMessage) chatCompletionRequest {
	out := make([]chatCompletionMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, chatCompletionMessage{Role: m.Role, Content: strings.TrimSpace(m.Content)})
	}
	return chatCompletionRequest{
		Model:       model,
		Messages:    out,
		Temperature: 0.9,
		MaxTokens:   maxReplyTokens,
	}
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// replyTextOf tolerates string content; array-part content is out of scope.
func replyTextOf(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
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

// callProvider sends one chat completion through the configured OpenAI-
// compatible endpoint (9Router /v1/chat/completions or Gemini's own OpenAI
// endpoint). apiKey travels only in the Authorization header, never in logs.
func (h *ChatHandler) callProvider(ctx context.Context, baseURL, apiKey, model string, body chatCompletionRequest) (string, int, int, error) {
	url := baseURL + "/chat/completions"
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
		req.Header.Set("Authorization", "Bearer "+apiKey)
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
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", 0, 0, fmt.Errorf("provider rate limited (429)")
		}
		if resp.StatusCode != http.StatusOK {
			return "", 0, 0, fmt.Errorf("provider status %d", resp.StatusCode)
		}
		var out chatCompletionResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			return "", 0, 0, err
		}
		if out.Error != nil {
			return "", 0, 0, fmt.Errorf("provider error: %s", out.Error.Message)
		}
		text := ""
		if len(out.Choices) > 0 {
			text = replyTextOf(out.Choices[0].Message.Content)
		}
		if text == "" {
			return "", 0, 0, fmt.Errorf("empty provider reply")
		}
		tIn, tOut := 0, 0
		if out.Usage != nil {
			tIn, tOut = out.Usage.PromptTokens, out.Usage.CompletionTokens
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
	if h.Cfg.ProviderAPIKey() == "" {
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
	model := h.Cfg.ProviderModel()
	reply, tIn, tOut, err := h.callProvider(ctx, h.Cfg.ProviderBaseURL(), h.Cfg.ProviderAPIKey(), model, toChatCompletionBody(model, req.Messages))

	row := models.ChatMessage{
		DeviceID:   deviceID,
		AppsName:   appsName,
		IdolName:   idolName,
		Provider:   "openai-compat:" + h.Cfg.ProviderModel(),
		TokensIn:   tIn,
		TokensOut:  tOut,
		Ok:         err == nil,
		OccurredAt: time.Now().UTC(),
		UsageDate:  today,
	}
	_ = h.DB.Create(&row).Error // logging must not fail the reply

	if err != nil {
		// Log status/error saja (tanpa key, tanpa isi pesan user) agar
		// diagnosis 502 tidak buta seperti sebelumnya.
		log.Printf("[chat] provider failed model=%s err=%v", h.Cfg.ProviderModel(), err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "provider unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply, "provider": "openai-compat:" + model, "model": model})
}

type chatIdolStat struct {
	IdolName      string `json:"idol_name"`
	Hits          int    `json:"hits"`
	UniqueDevices int    `json:"unique_devices"`
}

// DailyStats returns chat relay usage for a date (YYYY-MM-DD, Asia/Jakarta):
// total messages, unique devices, ok/failed split, token sums, 24 hourly
// buckets, per-app and per-idol breakdowns. Admin only (route-guarded).
func (h *ChatHandler) DailyStats(c *gin.Context) {
	dateStr := strings.TrimSpace(c.Query("date"))
	appsName := strings.ToLower(strings.TrimSpace(c.Query("apps_name")))

	if dateStr == "" {
		dateStr = jakartaNow().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	base := `FROM chat_messages WHERE usage_date = ?`
	args := []any{dateStr}
	if appsName != "" {
		base += ` AND apps_name = ?`
		args = append(args, appsName)
	}

	type totalRow struct {
		Hits        int
		DAU         int
		OkCount     int
		FailedCount int
		TokensIn    int
		TokensOut   int
	}
	var totals totalRow
	totalSQL := `SELECT COUNT(*) AS hits, COUNT(DISTINCT device_id) AS dau,
		COUNT(*) FILTER (WHERE ok) AS ok_count,
		COUNT(*) FILTER (WHERE NOT ok) AS failed_count,
		COALESCE(SUM(tokens_in), 0) AS tokens_in,
		COALESCE(SUM(tokens_out), 0) AS tokens_out ` + base
	if err := h.DB.Raw(totalSQL, args...).Scan(&totals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch totals"})
		return
	}

	type hourRow struct {
		Hour          int
		Hits          int
		UniqueDevices int
	}
	var hourRows []hourRow
	hourSQL := `SELECT EXTRACT(HOUR FROM occurred_at AT TIME ZONE 'Asia/Jakarta')::int AS hour,
		COUNT(*) AS hits, COUNT(DISTINCT device_id) AS unique_devices ` +
		base + ` GROUP BY 1 ORDER BY 1`
	if err := h.DB.Raw(hourSQL, args...).Scan(&hourRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch hourly"})
		return
	}
	byHour := make(map[int]hourRow, len(hourRows))
	for _, r := range hourRows {
		byHour[r.Hour] = r
	}
	hourly := make([]hourlyBucket, 0, 24)
	for hr := 0; hr < 24; hr++ {
		r := byHour[hr]
		hourly = append(hourly, hourlyBucket{Hour: hr, Hits: r.Hits, UniqueDevices: r.UniqueDevices})
	}

	var byApp []appStat
	appSQL := `SELECT apps_name AS apps_name, COUNT(*) AS hits,
		COUNT(DISTINCT device_id) AS unique_devices ` + base + ` GROUP BY 1 ORDER BY hits DESC LIMIT 20`
	if err := h.DB.Raw(appSQL, args...).Scan(&byApp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch apps"})
		return
	}
	if byApp == nil {
		byApp = []appStat{}
	}

	var byIdol []chatIdolStat
	idolSQL := `SELECT COALESCE(NULLIF(idol_name, ''), '(unknown)') AS idol_name, COUNT(*) AS hits,
		COUNT(DISTINCT device_id) AS unique_devices ` + base + `
		GROUP BY 1 ORDER BY hits DESC LIMIT 20`
	if err := h.DB.Raw(idolSQL, args...).Scan(&byIdol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch idols"})
		return
	}
	if byIdol == nil {
		byIdol = []chatIdolStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":           dateStr,
		"timezone":       "Asia/Jakarta",
		"apps_name":      appsName,
		"total_messages": totals.Hits,
		"unique_devices": totals.DAU,
		"ok_count":       totals.OkCount,
		"failed_count":   totals.FailedCount,
		"tokens_in":      totals.TokensIn,
		"tokens_out":     totals.TokensOut,
		"hourly":         hourly,
		"by_app":         byApp,
		"by_idol":        byIdol,
	})
}
