package utils

import (
	"context"
	"errors"
	"time"

	"google.golang.org/api/idtoken"
)

// GoogleProfile is the verified identity from a Google ID token.
type GoogleProfile struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

// VerifyGoogleIDToken validates a Google ID token from a mobile device.
// clientIDs are the allowed audiences (Android + iOS client IDs).
func VerifyGoogleIDToken(ctx context.Context, raw string, clientIDs ...string) (*GoogleProfile, error) {
	if raw == "" {
		return nil, errors.New("missing id_token")
	}
	// Try each configured audience; first success wins.
	var lastErr error
	for _, aud := range clientIDs {
		if aud == "" {
			continue
		}
		payload, err := idtoken.Validate(ctx, raw, aud)
		if err != nil {
			lastErr = err
			continue
		}
		profile := &GoogleProfile{}
		if v, ok := payload.Claims["sub"].(string); ok {
			profile.Sub = v
		}
		if v, ok := payload.Claims["email"].(string); ok {
			profile.Email = v
		}
		if v, ok := payload.Claims["name"].(string); ok {
			profile.Name = v
		}
		if v, ok := payload.Claims["picture"].(string); ok {
			profile.Picture = v
		}
		if v, ok := payload.Claims["email_verified"].(bool); ok && !v {
			return nil, errors.New("google email not verified")
		}
		if profile.Sub == "" || profile.Email == "" {
			return nil, errors.New("incomplete google profile")
		}
		if exp, ok := payload.Claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return nil, errors.New("google id_token expired")
			}
		}
		return profile, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("no google client id configured")
}
