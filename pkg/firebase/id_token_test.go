package firebase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testProject = "lodify-production"

func newTestSigner(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	certPem := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	return key, certPem
}

func serveCerts(t *testing.T, certs map[string]string) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_ = json.NewEncoder(w).Encode(certs)
	}))
	t.Cleanup(srv.Close)

	oldURL := googleCertsURL
	googleCertsURL = srv.URL
	t.Cleanup(func() {
		googleCertsURL = oldURL
		certsMu.Lock()
		certsKeys, certsExpiresAt, certsFetchedAt = nil, time.Time{}, time.Time{}
		certsMu.Unlock()
	})
}

func validClaims() jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{
		"iss":          "https://securetoken.google.com/" + testProject,
		"aud":          testProject,
		"sub":          "uid-1",
		"iat":          now.Unix(),
		"exp":          now.Add(time.Hour).Unix(),
		"auth_time":    now.Unix(),
		"phone_number": "+10000000000",
		"firebase":     map[string]any{"sign_in_provider": "phone"},
	}
}

func sign(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestVerifyPhoneIDToken(t *testing.T) {
	key, certPem := newTestSigner(t)
	otherKey, _ := newTestSigner(t)
	serveCerts(t, map[string]string{"k1": certPem})

	ctx := context.Background()

	tests := []struct {
		name   string
		token  func() string
		wantOK bool
	}{
		{"valid", func() string { return sign(t, key, "k1", validClaims()) }, true},
		{"wrong audience", func() string {
			c := validClaims()
			c["aud"] = "other-project"
			return sign(t, key, "k1", c)
		}, false},
		{"wrong issuer", func() string {
			c := validClaims()
			c["iss"] = "https://securetoken.google.com/other-project"
			return sign(t, key, "k1", c)
		}, false},
		{"expired", func() string {
			c := validClaims()
			c["exp"] = time.Now().Add(-2 * time.Minute).Unix()
			return sign(t, key, "k1", c)
		}, false},
		{"stale sign-in", func() string {
			c := validClaims()
			c["auth_time"] = time.Now().Add(-idTokenMaxAge - time.Minute).Unix()
			return sign(t, key, "k1", c)
		}, false},
		{"not a phone sign-in", func() string {
			c := validClaims()
			c["firebase"] = map[string]any{"sign_in_provider": "password"}
			return sign(t, key, "k1", c)
		}, false},
		{"missing iat", func() string {
			c := validClaims()
			delete(c, "iat")
			return sign(t, key, "k1", c)
		}, false},
		{"no phone number", func() string {
			c := validClaims()
			delete(c, "phone_number")
			return sign(t, key, "k1", c)
		}, false},
		{"signed by another key", func() string { return sign(t, otherKey, "k1", validClaims()) }, false},
		{"unknown kid", func() string { return sign(t, key, "k2", validClaims()) }, false},
		{"HS256 alg confusion", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
			tok.Header["kid"] = "k1"
			s, _ := tok.SignedString([]byte(certPem))
			return s
		}, false},
		{"tampered payload", func() string {
			parts := strings.Split(sign(t, key, "k1", validClaims()), ".")
			other := strings.Split(sign(t, key, "k1", jwt.MapClaims{"aud": testProject}), ".")
			return parts[0] + "." + other[1] + "." + parts[2]
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPhoneIDToken(ctx, testProject, tt.token())
			if (err == nil) != tt.wantOK {
				t.Fatalf("wantOK=%v, got err=%v", tt.wantOK, err)
			}
		})
	}
}

func TestCertFetchFailureIsThrottled(t *testing.T) {
	key, _ := newTestSigner(t)

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	serveCerts(t, nil)
	googleCertsURL = srv.URL

	for i := 0; i < 5; i++ {
		if err := VerifyPhoneIDToken(context.Background(), testProject, sign(t, key, "k1", validClaims())); err == nil {
			t.Fatal("expected failure while certs are unavailable")
		}
	}

	if n := hits.Load(); n != 1 {
		t.Fatalf("expected 1 fetch attempt within the throttle window, got %d", n)
	}
}

func TestLooksLikeIDToken(t *testing.T) {
	key, _ := newTestSigner(t)

	if !LooksLikeIDToken(sign(t, key, "k1", validClaims())) {
		t.Fatal("RS256 JWT with kid should look like an ID token")
	}

	for _, s := range []string{
		"",
		"AD8T5IsRoyBsUHQ5x0VvsXpJYZDa-opaque-session-info",
		"a.b.c",
		"a.b",
	} {
		if LooksLikeIDToken(s) {
			t.Fatalf("%q should not look like an ID token", s)
		}
	}
}
