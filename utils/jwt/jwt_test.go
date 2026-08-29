package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func generateTestCerts(t *testing.T) map[string]Cert {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicPEM, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicPEM})

	return map[string]Cert{
		"kid-1": {
			PrivateKey: privatePEM,
			PublicKey:  publicPEMBytes,
		},
	}
}

func TestJWTParseRejectsExpiredToken(t *testing.T) {
	certs := generateTestCerts(t)
	jwter := NewJwt(certs)

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.RegisteredClaims{
		Subject:   "alice",
		ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(-time.Minute)),
		IssuedAt:  jwtv5.NewNumericDate(time.Now().Add(-2 * time.Minute)),
	})
	token.Header["kid"] = "kid-1"

	str, err := jwter.Sign("kid-1", token)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = jwter.Parse(str, &jwtv5.RegisteredClaims{})
	if err == nil {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestJWTParseAcceptsFreshToken(t *testing.T) {
	certs := generateTestCerts(t)
	jwter := NewJwt(certs)

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.RegisteredClaims{
		Subject:   "alice",
		ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt:  jwtv5.NewNumericDate(time.Now()),
	})
	token.Header["kid"] = "kid-1"

	str, err := jwter.Sign("kid-1", token)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	parsed, err := jwter.Parse(str, &jwtv5.RegisteredClaims{})
	if err != nil {
		t.Fatalf("parse valid token: %v", err)
	}
	if parsed == nil || parsed.Claims == nil {
		t.Fatal("parsed token or claims is nil")
	}
}
