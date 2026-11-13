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
	"sync"
	"sync/atomic"
	"time"

	"admin_daebak_music/backend/internal/config"
	"admin_daebak_music/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Chat relay: mobile POSTs here. Server memegang key 9Router (tidak pernah
// dikirim di APK). SEMUA chat wajib lewat 9Router (OpenAI-compatible
// /v1/chat/completions) — tidak ada jalur langsung ke Gemini. Limit harian
// diatur di mobile apps, bukan backend. Backend TIDAK mem-pin satu model:
// daftar model chat diambil otomatis dari 9Router (GET /v1/models, di-cache
// singkat) lalu dipilih round-robin per request dengan failover ke model
// berikut bila gagal, sehingga seluruh model/provider yang tersedia di
// 9Router ikut terpakai. Contract:
// success -> {"reply","provider","model","requested_model"},
// provider failure -> 502.

type ChatHandler struct {
	DB  *gorm.DB
	Cfg *config.Config

	// Cache daftar model 9Router + counter round-robin. Diakses dari
	// banyak goroutine (satu per request) sehingga butuh mutex/atomic.
	modelsMu        sync.Mutex
	cachedModels    []string
	modelsFetchedAt time.Time
	rrCounter       atomic.Uint64
}

// nineRouterModelsTTL: berapa lama daftar model 9Router di-cache sebelum
// diambil ulang. Singkat agar model/provider baru di UI 9Router cepat
// ikut terpakai, tapi tidak membebani tiap request dengan GET /v1/models.
const nineRouterModelsTTL = 5 * time.Minute

type nineRouterModelEntry struct {
	ID string `json:"id"`
}

type nineRouterModelList struct {
	Data []nineRouterModelEntry `json:"data"`
}

// fetchModels mengambil daftar id model chat dari 9Router (GET /v1/models
// = endpoint daftar model chat/LLM; combo ikut muncul sebagai id biasa).
func (h *ChatHandler) fetchModels(ctx context.Context) ([]string, error) {
	url := strings.TrimRight(h.Cfg.ProviderBaseURL(), "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+h.Cfg.ProviderAPIKey())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models status %d", resp.StatusCode)
	}
	var out nineRouterModelList
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return nil, err
	}
	var ids []string
	seen := map[string]bool{}
	for _, m := range out.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("empty model list from 9Router")
	}
	return ids, nil
}

// modelsSnapshot mengembalikan daftar model 9Router (cache singkat,
// fallback ke cache basi bila fetch gagal). Dipakai pickModel dan failover.
func (h *ChatHandler) modelsSnapshot(ctx context.Context) ([]string, error) {
	h.modelsMu.Lock()
	fresh := len(h.cachedModels) > 0 && time.Since(h.modelsFetchedAt) < nineRouterModelsTTL
	cached := h.cachedModels
	h.modelsMu.Unlock()

	if fresh {
		return cached, nil
	}
	fetched, err := h.fetchModels(ctx)
	if err != nil {
		if len(cached) == 0 {
			return nil, err
		}
		log.Printf("[chat] models refresh failed, using %d stale models: %v", len(cached), err)
		return cached, nil
	}
	h.modelsMu.Lock()
	h.cachedModels = fetched
	h.modelsFetchedAt = time.Now()
	h.modelsMu.Unlock()
	return fetched, nil
}

// pickModel memilih satu model dari daftar 9Router secara round-robin.
// Daftar di-cache singkat; bila fetch gagal, cache basi dipakai supaya
// request tetap jalan selama 9Router sempat merespons sebelumnya.
func (h *ChatHandler) pickModel(ctx context.Context) (string, error) {
	list, err := h.modelsSnapshot(ctx)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("no models available from 9Router")
	}
	n := h.rrCounter.Add(1)
	return list[int(n-1)%len(list)], nil
}

// maxChatModelAttempts: berapa model berbeda dicoba per request sebelum
// menyerah (502). Satu model rusak tidak menggagalkan request selama ada
// model sehat lain di 9Router.
const maxChatModelAttempts = 3

// chatWithFailover mengirim chat via 9Router, mulai dari model hasil
// round-robin lalu lanjut ke model berikut bila gagal (400/429/5xx/
// network). Mengembalikan requested = model yang akhirnya dipakai (atau
// terakhir dicoba) dan served = model aktual dari respons upstream.
func (h *ChatHandler) chatWithFailover(ctx context.Context, baseURL, apiKey string, msgs []chatIncomingMessage) (reply string, tIn, tOut int, requested, served string, err error) {
	list, err := h.modelsSnapshot(ctx)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	if len(list) == 0 {
		return "", 0, 0, "", "", fmt.Errorf("no models available from 9Router")
	}
	start := int(h.rrCounter.Add(1)-1) % len(list)
	attempts := len(list)
	if attempts > maxChatModelAttempts {
		attempts = maxChatModelAttempts
	}
	tried := map[string]bool{}
	for i := 0; i < len(list) && len(tried) < attempts; i++ {
		m := list[(start+i)%len(list)]
		if tried[m] {
			continue
		}
		tried[m] = true
		requested = m
		var e error
		reply, tIn, tOut, served, e = h.callProvider(ctx, baseURL, apiKey, m, toChatCompletionBody(m, msgs))
		if e == nil {
			return reply, tIn, tOut, requested, served, nil
		}
		err = e
		served = m // gagal sebelum respons terbaca — catat yang diminta
		log.Printf("[chat] model attempt failed requested=%s err=%v, trying next", m, e)
	}
	return "", 0, 0, requested, served, err
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
// shape served by 9Router (/v1/chat/completions). Roles pass through
// (system/user/assistant). Pure (unit-tested).
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
	Model   string `json:"model"`
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

// callProvider sends one chat completion through 9Router
// (/v1/chat/completions, OpenAI-compatible). apiKey travels only in the
// Authorization header, never in logs.
func (h *ChatHandler) callProvider(ctx context.Context, baseURL, apiKey, model string, body chatCompletionRequest) (text string, tIn, tOut int, servedModel string, err error) {
	url := baseURL + "/chat/completions"
	payload, err := json.Marshal(body)
	if err != nil {
		return "", 0, 0, "", err
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", 0, 0, "", ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return "", 0, 0, "", err
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
			return "", 0, 0, "", fmt.Errorf("provider rate limited (429)")
		}
		if resp.StatusCode != http.StatusOK {
			return "", 0, 0, "", fmt.Errorf("provider status %d", resp.StatusCode)
		}
		var out chatCompletionResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			// Sertakan cuplikan body (200 char) agar respons non-JSON
			// upstream (mis. SSE/halaman error) bisa didiagnosis dari log.
			snip := strings.TrimSpace(string(raw))
			if len(snip) > 200 {
				snip = snip[:200]
			}
			return "", 0, 0, "", fmt.Errorf("decode provider reply: %v (body %.200q)", err, snip)
		}
		if out.Error != nil {
			return "", 0, 0, "", fmt.Errorf("provider error: %s", out.Error.Message)
		}
		text := ""
		if len(out.Choices) > 0 {
			text = replyTextOf(out.Choices[0].Message.Content)
		}
		if text == "" {
			return "", 0, 0, "", fmt.Errorf("empty provider reply")
		}
		tIn, tOut := 0, 0
		if out.Usage != nil {
			tIn, tOut = out.Usage.PromptTokens, out.Usage.CompletionTokens
		}
		// Nama model yang benar-benar menjawab (bisa beda dari yang diminta
		// bila gateway failover, mis. 9Router → OpenCode Free).
		served := strings.TrimSpace(out.Model)
		if served == "" {
			served = model
		}
		return text, tIn, tOut, served, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("provider unavailable")
	}
	return "", 0, 0, "", lastErr
}

// PostChat relays one chat message via 9Router using the server-held key.
// Limit harian diatur di mobile apps, bukan backend: setiap request valid
// langsung diteruskan (tetap dicatat di DB untuk monitoring/statistik).
// Model dipilih otomatis round-robin dari daftar /v1/models 9Router per
// request, dengan failover ke model berikut (maks 3) bila gagal — satu
// model rusak tidak menggagalkan request selama ada model sehat lain.
// requested = model yang akhirnya dipakai; served = model aktual dari
// respons upstream. Keduanya dicatat di log dan respons agar transparan.
func (h *ChatHandler) PostChat(c *gin.Context) {
	if !h.Cfg.ChatRelayReady() {
		log.Printf("[chat] relay disabled: CHAT_PROVIDER_API_KEY empty")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat relay disabled (9Router not configured)"})
		return
	}
	baseURL := h.Cfg.ProviderBaseURL()
	if isDirectGeminiURL(baseURL) {
		// Pertahanan berlapis: backend tidak boleh menembak langsung ke
		// Gemini — semua chat wajib lewat 9Router.
		log.Printf("[chat] misconfigured base_url=%s: direct Gemini blocked, use 9Router", baseURL)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "chat relay misconfigured (direct Gemini blocked, use 9Router)"})
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
	today := jakartaNow().Format("2006-01-02")

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.Cfg.ChatTimeoutSec)*time.Second)
	defer cancel()
	// Tanpa model tetap: pilih otomatis dari daftar model 9Router, dengan
	// failover ke model berikut bila yang pertama gagal.
	// 502 agar mobile fallback, konsisten dengan kegagalan provider lain.
	started := time.Now()
	reply, tIn, tOut, model, served, err := h.chatWithFailover(ctx, baseURL, h.Cfg.ProviderAPIKey(), req.Messages)
	if served == "" {
		served = model // gagal sebelum respons terbaca — catat yang diminta
	}
	providerLabel := "9router:" + served

	row := models.ChatMessage{
		DeviceID:   deviceID,
		AppsName:   appsName,
		IdolName:   idolName,
		Provider:   providerLabel,
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
		if model == "" {
			// Daftar model pun tak didapat (9Router down/key salah/kosong).
			log.Printf("[chat] no models available base_url=%s err=%v", baseURL, err)
		} else {
			log.Printf("[chat] provider failed base_url=%s requested=%s served=%s err=%v", baseURL, model, served, err)
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "provider unavailable"})
		return
	}
	log.Printf("[chat] ok base_url=%s requested=%s served=%s tokens_in=%d tokens_out=%d latency_ms=%d",
		baseURL, model, served, tIn, tOut, time.Since(started).Milliseconds())
	c.JSON(http.StatusOK, gin.H{"reply": reply, "provider": providerLabel, "model": served, "requested_model": model})
}

// isDirectGeminiURL true bila base URL menunjuk langsung ke Gemini, bukan
// ke gateway 9Router. Dipakai sebagai guard agar tidak ada jalur langsung.
func isDirectGeminiURL(baseURL string) bool {
	u := strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(u, "generativelanguage.googleapis.com")
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

	type providerStat struct {
		Provider      string `json:"provider"`
		Hits          int    `json:"hits"`
		UniqueDevices int    `json:"unique_devices"`
	}
	var byProvider []providerStat
	providerSQL := `SELECT COALESCE(NULLIF(provider, ''), '(unknown)') AS provider, COUNT(*) AS hits,
		COUNT(DISTINCT device_id) AS unique_devices ` + base + `
		GROUP BY 1 ORDER BY hits DESC`
	if err := h.DB.Raw(providerSQL, args...).Scan(&byProvider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch providers"})
		return
	}
	if byProvider == nil {
		byProvider = []providerStat{}
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
		"by_provider":    byProvider,
	})
}
