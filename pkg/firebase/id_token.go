package firebase

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// idTokenMaxAge bounds how long after the phone sign-in an ID token is
	// still accepted, so a leaked token can't be replayed for its full hour.
	idTokenMaxAge = 10 * time.Minute

	certsRefetchInterval = time.Minute
	certsDefaultTTL      = time.Hour
)

// googleCertsURL serves the x509 certificates that sign Firebase ID tokens.
var googleCertsURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

var (
	certsMu        sync.Mutex
	certsKeys      map[string]*rsa.PublicKey
	certsExpiresAt time.Time
	certsFetchedAt time.Time
	certsClient    = &http.Client{Timeout: 10 * time.Second}

	maxAgeRe = regexp.MustCompile(`max-age=(\d+)`)
)

// LooksLikeIDToken reports whether s is shaped like a Firebase ID token (an
// RS256 JWT with a key id), as opposed to the opaque sessionInfo returned by
// accounts:sendVerificationCode.
func LooksLikeIDToken(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err = json.Unmarshal(raw, &header); err != nil {
		return false
	}

	return header.Alg == jwt.SigningMethodRS256.Alg() && header.Kid != ""
}

// VerifyPhoneIDToken checks that idToken is a genuine Firebase ID token of
// firebaseProjectId, issued for a phone-number sign-in in the last
// idTokenMaxAge.
func VerifyPhoneIDToken(ctx context.Context, firebaseProjectId, idToken string) error {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(
		idToken,
		claims,
		func(t *jwt.Token) (any, error) {
			kid, _ := t.Header["kid"].(string)
			return publicKey(ctx, kid)
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithAudience(firebaseProjectId),
		jwt.WithIssuer("https://securetoken.google.com/"+firebaseProjectId),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}

	if sub, _ := claims["sub"].(string); sub == "" {
		return errors.New("failed to verify: empty subject")
	}

	if iat, _ := claims["iat"].(float64); iat <= 0 {
		return errors.New("failed to verify: missing iat")
	}

	authTime, ok := claims["auth_time"].(float64)
	if !ok {
		return errors.New("failed to verify: missing auth_time")
	}
	signedInAt := time.Unix(int64(authTime), 0)
	if time.Since(signedInAt) > idTokenMaxAge || time.Until(signedInAt) > time.Minute {
		return errors.New("failed to verify: sign-in is not recent")
	}

	if phone, _ := claims["phone_number"].(string); phone == "" {
		return errors.New("failed to verify: no phone number")
	}

	fb, _ := claims["firebase"].(map[string]any)
	if provider, _ := fb["sign_in_provider"].(string); provider != "phone" {
		return errors.New("failed to verify: not a phone sign-in")
	}

	return nil
}

func publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, errors.New("missing kid")
	}

	certsMu.Lock()
	defer certsMu.Unlock()

	now := time.Now()
	stale := now.After(certsExpiresAt)
	key, ok := certsKeys[kid]

	// Refetch when the cache expired or the kid is unknown (Google rotated
	// keys), but attempt it at most once a minute — failed attempts included —
	// so made-up kids or a Google outage can't make every request call Google
	// while holding the lock.
	if (stale || !ok) && now.Sub(certsFetchedAt) > certsRefetchInterval {
		certsFetchedAt = now
		if keys, ttl, err := fetchCerts(ctx); err == nil {
			certsKeys, certsExpiresAt = keys, now.Add(ttl)
			stale = false
			key, ok = certsKeys[kid]
		}
	}

	if stale {
		return nil, errors.New("signing keys unavailable")
	}
	if !ok {
		return nil, errors.New("unknown signing key")
	}

	return key, nil
}

func fetchCerts(ctx context.Context) (map[string]*rsa.PublicKey, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleCertsURL, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := certsClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("fetch firebase certs: status %d", resp.StatusCode)
	}

	var pems map[string]string
	if err = json.NewDecoder(resp.Body).Decode(&pems); err != nil {
		return nil, 0, err
	}

	keys := make(map[string]*rsa.PublicKey, len(pems))
	for kid, certPem := range pems {
		block, _ := pem.Decode([]byte(certPem))
		if block == nil {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		if pub, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			keys[kid] = pub
		}
	}
	if len(keys) == 0 {
		return nil, 0, errors.New("fetch firebase certs: no usable keys")
	}

	ttl := certsDefaultTTL
	if m := maxAgeRe.FindStringSubmatch(resp.Header.Get("Cache-Control")); m != nil {
		if secs, err := strconv.Atoi(m[1]); err == nil && secs > 0 {
			ttl = time.Duration(secs) * time.Second
		}
	}
	// Keys must outlive the refetch throttle, or they'd go unusable until the
	// next fetch is allowed.
	if ttl < certsRefetchInterval {
		ttl = certsRefetchInterval
	}

	return keys, ttl, nil
}
