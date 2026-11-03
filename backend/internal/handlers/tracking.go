package handlers

import (
	"net/http"
	"strings"
	"time"

	"admin_daebak_music/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrackingHandler struct {
	DB *gorm.DB
}

type trackDailyUseRequest struct {
	DeviceID       string `json:"device_id" binding:"required"`
	AppsName       string `json:"apps_name" binding:"required"`
	AndroidVersion string `json:"android_version"`
	DeviceBrand    string `json:"device_brand"`
	DeviceModel    string `json:"device_model"`
	Country        string `json:"country"`
	CountryCode    string `json:"country_code"`
	Email          string `json:"email"`
}

// parseCountryCodes accepts repeated ?country_code=ID&country_code=US and/or
// comma-separated ?country_code=ID,US. Returns uppercased, deduped codes.
func parseCountryCodes(raw []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, r := range raw {
		for _, part := range strings.Split(r, ",") {
			cc := strings.ToUpper(strings.TrimSpace(part))
			if cc == "" {
				continue
			}
			if _, ok := seen[cc]; !ok {
				seen[cc] = struct{}{}
				out = append(out, cc)
			}
		}
	}
	return out
}

// parseApps accepts repeated/comma-separated ?apps_name=. Lowercased, deduped.
func parseApps(raw []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, r := range raw {
		for _, part := range strings.Split(r, ",") {
			app := strings.ToLower(strings.TrimSpace(part))
			if app == "" {
				continue
			}
			if _, ok := seen[app]; !ok {
				seen[app] = struct{}{}
				out = append(out, app)
			}
		}
	}
	return out
}

var jakartaLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	jakartaLoc = loc
}

func jakartaNow() time.Time {
	return time.Now().In(jakartaLoc)
}

// TrackDailyUse records one anonymous usage hit. Every call inserts a row.
func (h *TrackingHandler) TrackDailyUse(c *gin.Context) {
	var req trackDailyUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	appsName := strings.ToLower(strings.TrimSpace(req.AppsName))
	androidVersion := strings.TrimSpace(req.AndroidVersion)
	deviceBrand := strings.TrimSpace(req.DeviceBrand)
	deviceModel := strings.TrimSpace(req.DeviceModel)
	country := strings.TrimSpace(req.Country)
	countryCode := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	email := strings.TrimSpace(req.Email)

	if deviceID == "" || appsName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id and apps_name are required"})
		return
	}
	if len(country) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country max 100 characters"})
		return
	}
	if len(countryCode) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country_code max 10 characters"})
		return
	}
	if len(deviceBrand) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_brand max 100 characters"})
		return
	}
	if len(deviceModel) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_model max 100 characters"})
		return
	}
	if email != "" && !strings.Contains(email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	now := jakartaNow()
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	ev := models.AppUsageEvent{
		DeviceID:       deviceID,
		AppsName:       appsName,
		AndroidVersion: androidVersion,
		DeviceBrand:    deviceBrand,
		DeviceModel:    deviceModel,
		Country:        country,
		CountryCode:    countryCode,
		Email:          emailPtr,
		OccurredAt:     time.Now().UTC(),
		UsageDate:      now.Format("2006-01-02"),
	}
	if err := h.DB.Create(&ev).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record usage"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "recorded", "event": ev})
}

type hourlyBucket struct {
	Hour           int `json:"hour"`
	Hits           int `json:"hits"`
	UniqueDevices  int `json:"unique_devices"`
}

type countryStat struct {
	Country       string `json:"country"`
	CountryCode   string `json:"country_code"`
	Hits          int    `json:"hits"`
	UniqueDevices int    `json:"unique_devices"`
}

type optionCountry struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

type appStat struct {
	AppsName      string `json:"apps_name"`
	Hits          int    `json:"hits"`
	UniqueDevices int    `json:"unique_devices"`
}

// DailyStats returns DAU (distinct device_id), total hits, 24 hourly buckets
// and top countries for a given date (YYYY-MM-DD, Asia/Jakarta).
func (h *TrackingHandler) DailyStats(c *gin.Context) {
	dateStr := strings.TrimSpace(c.Query("date"))
	appsNames := parseApps(c.QueryArray("apps_name"))
	countryCodes := parseCountryCodes(c.QueryArray("country_code"))

	if dateStr == "" {
		dateStr = jakartaNow().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	// Optional app scope (multi-select). Lowercased on write.
	applyApps := func(sql *string, args *[]any) {
		if len(appsNames) == 0 {
			return
		}
		placeholders := make([]string, len(appsNames))
		for i, app := range appsNames {
			placeholders[i] = "?"
			*args = append(*args, app)
		}
		*sql += ` AND apps_name IN (` + strings.Join(placeholders, ",") + `)`
	}

	// Optional country scope (multi-select). Matches country_code first,
	// falls back to legacy rows that only have country (= code, uppercased).
	applyCountry := func(sql *string, args *[]any) {
		if len(countryCodes) == 0 {
			return
		}
		placeholders := make([]string, len(countryCodes))
		for i, cc := range countryCodes {
			placeholders[i] = "?"
			*args = append(*args, cc)
		}
		*sql += ` AND COALESCE(NULLIF(UPPER(country_code), ''), NULLIF(UPPER(country), '')) IN (` +
			strings.Join(placeholders, ",") + `)`
	}

	// Hourly buckets in WIB. Use Raw so GROUP BY 1 positional isn't quoted by GORM.
	type hourRow struct {
		Hour          int
		Hits          int
		UniqueDevices int
	}
	var hourRows []hourRow
	hourSQL := `SELECT EXTRACT(HOUR FROM occurred_at AT TIME ZONE 'Asia/Jakarta')::int AS hour, COUNT(*) AS hits, COUNT(DISTINCT device_id) AS unique_devices
		FROM daily_app_usages WHERE usage_date = ?`
	hourArgs := []any{dateStr}
	applyApps(&hourSQL, &hourArgs)
	applyCountry(&hourSQL, &hourArgs)
	hourSQL += ` GROUP BY 1 ORDER BY 1`
	if err := h.DB.Raw(hourSQL, hourArgs...).Scan(&hourRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch hourly stats: " + err.Error()})
		return
	}
	byHour := make(map[int]hourRow, len(hourRows))
	for _, r := range hourRows {
		byHour[r.Hour] = r
	}
	hourly := make([]hourlyBucket, 0, 24)
	var totalHits, dau int
	for hr := 0; hr < 24; hr++ {
		r := byHour[hr]
		hourly = append(hourly, hourlyBucket{Hour: hr, Hits: r.Hits, UniqueDevices: r.UniqueDevices})
		totalHits += r.Hits
	}

	// DAU + total hits overall for the date.
	type totalRow struct {
		Hits int
		DAU  int
	}
	var totals totalRow
	totalSQL := `SELECT COUNT(*) AS hits, COUNT(DISTINCT device_id) AS dau FROM daily_app_usages WHERE usage_date = ?`
	totalArgs := []any{dateStr}
	applyApps(&totalSQL, &totalArgs)
	applyCountry(&totalSQL, &totalArgs)
	if err := h.DB.Raw(totalSQL, totalArgs...).Scan(&totals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch totals: " + err.Error()})
		return
	}
	totalHits = totals.Hits
	dau = totals.DAU

	// Top countries grouped by country_code (fallback to country name for legacy rows).
	// Use Raw for the same quoting reason.
	var countries []countryStat
	countrySQL := `SELECT
			COALESCE(NULLIF(UPPER(country_code), ''), NULLIF(UPPER(country), ''), 'UNKNOWN') AS country_code,
			COALESCE(NULLIF(country, ''), NULLIF(country_code, ''), 'UNKNOWN') AS country,
			COUNT(*) AS hits, COUNT(DISTINCT device_id) AS unique_devices
		FROM daily_app_usages WHERE usage_date = ?`
	countryArgs := []any{dateStr}
	applyApps(&countrySQL, &countryArgs)
	applyCountry(&countrySQL, &countryArgs)
	countrySQL += ` GROUP BY 1, 2 ORDER BY hits DESC LIMIT 10`
	if err := h.DB.Raw(countrySQL, countryArgs...).Scan(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch country stats: " + err.Error()})
		return
	}
	if countries == nil {
		countries = []countryStat{}
	}

	// Per-app breakdown. Respects the country scope (global slice); when apps
	// are selected it shows only those, otherwise all apps for comparison.
	var byApp []appStat
	appSQL := `SELECT apps_name AS apps_name, COUNT(*) AS hits, COUNT(DISTINCT device_id) AS unique_devices
		FROM daily_app_usages WHERE usage_date = ?`
	appArgs := []any{dateStr}
	applyApps(&appSQL, &appArgs)
	applyCountry(&appSQL, &appArgs)
	appSQL += ` GROUP BY 1 ORDER BY hits DESC LIMIT 20`
	if err := h.DB.Raw(appSQL, appArgs...).Scan(&byApp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch apps stats: " + err.Error()})
		return
	}
	if byApp == nil {
		byApp = []appStat{}
	}

	// Option lists for the filter modal. Scoped by date only (ignores the
	// active apps/country filters) so the pill lists always match the
	// selected date and never collapse to the filtered subset.
	var allApps []string
	allAppSQL := `SELECT apps_name AS apps_name FROM daily_app_usages WHERE usage_date = ? GROUP BY 1 ORDER BY 1`
	if err := h.DB.Raw(allAppSQL, dateStr).Scan(&allApps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch apps options: " + err.Error()})
		return
	}
	if allApps == nil {
		allApps = []string{}
	}
	var allCountries []countryStat
	allCountrySQL := `SELECT
			COALESCE(NULLIF(UPPER(country_code), ''), NULLIF(UPPER(country), ''), 'UNKNOWN') AS country_code,
			COALESCE(NULLIF(country, ''), NULLIF(country_code, ''), 'UNKNOWN') AS country,
			COUNT(*) AS hits, COUNT(DISTINCT device_id) AS unique_devices
		FROM daily_app_usages WHERE usage_date = ? GROUP BY 1, 2 ORDER BY 2`
	if err := h.DB.Raw(allCountrySQL, dateStr).Scan(&allCountries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch country options: " + err.Error()})
		return
	}
	if allCountries == nil {
		allCountries = []countryStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":         dateStr,
		"timezone":     "Asia/Jakarta",
		"apps_name":    strings.Join(appsNames, ","),
		"country_code": strings.Join(countryCodes, ","),
		"total_hits":  totalHits,
		"dau":         dau,
		"hourly":      hourly,
		"by_country":  countries,
		"by_app":      byApp,
		"all_apps":    allApps,
		"all_countries": allCountries,
	})
}

// FilterOptions returns distinct apps and countries that have hits on a given
// date (Asia/Jakarta). Lightweight source for the dashboard filter modal pills,
// so the lists always follow the selected date.
func (h *TrackingHandler) FilterOptions(c *gin.Context) {
	dateStr := strings.TrimSpace(c.Query("date"))
	if dateStr == "" {
		dateStr = jakartaNow().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	var apps []string
	if err := h.DB.Raw(`SELECT apps_name AS apps_name FROM daily_app_usages
		WHERE usage_date = ? GROUP BY 1 ORDER BY 1`, dateStr).Scan(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch apps options: " + err.Error()})
		return
	}
	if apps == nil {
		apps = []string{}
	}
	var countries []optionCountry
	if err := h.DB.Raw(`SELECT
			COALESCE(NULLIF(UPPER(country_code), ''), NULLIF(UPPER(country), ''), 'UNKNOWN') AS country_code,
			COALESCE(NULLIF(country, ''), NULLIF(country_code, ''), 'UNKNOWN') AS country
		FROM daily_app_usages WHERE usage_date = ? GROUP BY 1, 2 ORDER BY 2`, dateStr).Scan(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch country options: " + err.Error()})
		return
	}
	if countries == nil {
		countries = []optionCountry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":      dateStr,
		"timezone":  "Asia/Jakarta",
		"apps":      apps,
		"countries": countries,
	})
}
