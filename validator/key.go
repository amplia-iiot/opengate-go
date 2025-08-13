package validator

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ErrUnsupportedSigningAlg string

func (e ErrUnsupportedSigningAlg) Error() string {
	return fmt.Sprintf("unsupported signing algorithm %v", string(e))
}

func unsupportedToken(token *jwt.Token) error {
	return ErrUnsupportedSigningAlg(token.Method.Alg())
}

type keys struct {
	hs256 []byte
	hs384 []byte
	hs512 []byte
	rs256 *rsa.PublicKey
	rs384 *rsa.PublicKey
	rs512 *rsa.PublicKey
}

type KeysConfig struct {
	HS256 string
	HS384 string
	HS512 string
	RS256 string
	RS384 string
	RS512 string
}

func NewKeys(config KeysConfig) (*keys, error) {
	k := keys{}
	if config.HS256 != "" {
		k.hs256 = []byte(config.HS256)
	}
	if config.HS384 != "" {
		k.hs384 = []byte(config.HS384)
	}
	if config.HS512 != "" {
		k.hs512 = []byte(config.HS512)
	}
	if config.RS256 != "" {
		rsa, err := getRsa(config.RS256)
		if err != nil {
			return nil, err
		}
		k.rs256 = rsa
	}
	if config.RS384 != "" {
		rsa, err := getRsa(config.RS384)
		if err != nil {
			return nil, err
		}
		k.rs384 = rsa
	}
	if config.RS512 != "" {
		rsa, err := getRsa(config.RS512)
		if err != nil {
			return nil, err
		}
		k.rs512 = rsa
	}
	return &k, nil
}

func MustNewKeys(config KeysConfig) *keys {
	k, err := NewKeys(config)
	if err != nil {
		panic(err)
	}
	return k
}

func NewRS256OrHS256Key(key string) *keys {
	k := keys{}
	if rsaKey, err := getRsa(key); err == nil {
		slog.Info("adding rsa key", "rsa", rsaKey)
		k.rs256 = rsaKey
	} else {
		slog.Info("adding hs key", "hs", key, "error", err)
		k.hs256 = []byte(key)
	}
	return &k
}

func getRsa(key string) (*rsa.PublicKey, error) {
	pubKeyPem := []byte(key)
	if !strings.Contains(key, "BEGIN PUBLIC KEY") {
		var err64 error
		pubKeyPem, err64 = base64.StdEncoding.DecodeString(key)
		if err64 != nil {
			return nil, fmt.Errorf("key is not a public key (pem / base64 pem): %w", err64)
		}
	}
	return jwt.ParseRSAPublicKeyFromPEM(pubKeyPem)
}

func (k *keys) Keyfunc(token *jwt.Token) (interface{}, error) {
	switch token.Method.Alg() {
	case jwt.SigningMethodHS256.Name:
		if len(k.hs256) == 0 {
			return nil, unsupportedToken(token)
		}
		return k.hs256, nil
	case jwt.SigningMethodHS384.Name:
		if len(k.hs384) == 0 {
			return nil, unsupportedToken(token)
		}
		return k.hs384, nil
	case jwt.SigningMethodHS512.Name:
		if len(k.hs512) == 0 {
			return nil, unsupportedToken(token)
		}
		return k.hs512, nil
	case jwt.SigningMethodRS256.Name:
		if k.rs256 == nil {
			return nil, unsupportedToken(token)
		}
		return k.rs256, nil
	case jwt.SigningMethodRS384.Name:
		if k.rs384 == nil {
			return nil, unsupportedToken(token)
		}
		return k.rs384, nil
	case jwt.SigningMethodRS512.Name:
		if k.rs512 == nil {
			return nil, unsupportedToken(token)
		}
		return k.rs512, nil
	default:
		return nil, unsupportedToken(token)
	}
}
