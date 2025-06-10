package jwt

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type Jwter interface {
	Parse(authToken string, claims jwt.Claims) (*jwt.Token, error)
	Sign(certKey string, token *jwt.Token) (tokenString string, err error)
}
type jwtCert struct {
	certs map[string]Cert
}
type Cert struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewJwt(certs map[string]Cert) Jwter {
	c := make(map[string]Cert, len(certs))
	for k, v := range certs {
		c[k] = v
	}
	return &jwtCert{certs: c}
}

func (j *jwtCert) Parse(authToken string, claims jwt.Claims) (*jwt.Token, error) {
	if authToken == "" {
		return nil, errors.New("authToken i empty string")
	}
	if j.certs == nil {
		return nil, errors.New("certs is empty")
	}
	// Parse token
	return jwt.ParseWithClaims(authToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		if token.Header == nil {
			return nil, errors.Errorf("Can't find the key: %v", token)
		}
		headerKid, ok := token.Header["kid"]
		if !ok {
			return nil, errors.Errorf("Can't find the key: %v", token)
		}
		kid, ok := headerKid.(string)
		if !ok {
			return nil, errors.Errorf("Can't find the key: %v", token)
		}
		if cert, ok := j.certs[kid]; ok {
			return jwt.ParseRSAPublicKeyFromPEM(cert.PublicKey)
		}
		return nil, errors.Errorf("Can't find the key: %v", token.Header["kid"])
	}, jwt.WithoutClaimsValidation())
}

func (j *jwtCert) Sign(certKey string, token *jwt.Token) (tokenString string, err error) {
	if token == nil {
		return "", errors.New("token is empty")
	}
	cert, isOk := j.certs[certKey]
	if !isOk {
		return "", errors.New("certs is empty")
	}
	var parsedKey *rsa.PrivateKey
	parsedKey, err = jwt.ParseRSAPrivateKeyFromPEM(cert.PrivateKey)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return token.SignedString(parsedKey)
}
