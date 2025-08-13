package validator

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type OpenGateClaims struct {
	ApiKey string `json:"X-ApiKey"`
	Name   string `json:"name"`
	//Sub    string `json:"sub"` included in RegisteredClaims as Subject field
	jwt.RegisteredClaims
}

const (
	HeaderApiKey  = "X-ApiKey"
	HeaderApiPass = "X-ApiPass"
	HeaderAuth    = "Authorization"
)

var (
	ErrNoAuth      = errors.New("no authorization method found in header")
	ErrParseClaims = errors.New("claims couldn't be parsed")
)

func getKeyFunc(privateKey string) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() == "RS256" {
			pubKeyPem := []byte(privateKey)
			if !strings.Contains(privateKey, "BEGIN PUBLIC KEY") {
				var err64 error
				pubKeyPem, err64 = base64.StdEncoding.DecodeString(privateKey)
				if err64 != nil {
					return nil, fmt.Errorf("privatekey is not a public key (pem / base64 pem): %w", err64)
				}
			}
			spkiBlock, _ := pem.Decode(pubKeyPem)
			var spkiKey *rsa.PublicKey
			pubInterface, _ := x509.ParsePKIXPublicKey(spkiBlock.Bytes)
			spkiKey = pubInterface.(*rsa.PublicKey)
			return spkiKey, nil
		}

		return []byte(privateKey), nil // using a config struct to handle the secret
	}
}

// GetApiKey returns the apiKey contained in the headers. In the case of a jwt
// it's validated with the privateKey.
//
// Deprecated: GetApiKey exists for historical compatibility and should not be
// used for performance reasons. To get the api key use GetApiKeyWithKeys or
// GetApiKeyWithKeyfunc, passing your desired implementation. There are keys
// implementations for the previous mode (NewRS256OrHS256Key) or a configurable
// one for all HS and RS algorithms (NewKeys).
func GetApiKey(headers http.Header, privateKey string) (apiKey string, err error) {
	return GetApiKeyWithKeys(headers, NewRS256OrHS256Key(privateKey))
}

// GetApiKeyWithKeys returns the apiKey contained in the headers. In the case of a jwt
// it's validated with the proper key based on the algorithm.
func GetApiKeyWithKeys(headers http.Header, keys *keys) (apiKey string, err error) {
	return GetApiKeyWithKeyfunc(headers, keys.Keyfunc)
}

// GetApiKeyWithKeys returns the apiKey contained in the headers. In the case of a jwt
// it's validated with the key returned by keyfunc.
func GetApiKeyWithKeyfunc(headers http.Header, keyfunc jwt.Keyfunc) (apiKey string, err error) {
	hApiKey := headers.Get(HeaderApiKey)
	//apiPass := headers.Get(HeaderApiPass)
	auth := headers.Get(HeaderAuth)

	if auth != "" {
		token := strings.Split(auth, "Bearer ")[1]
		if token != "" {
			var ogClaims OpenGateClaims
			_, err = jwt.ParseWithClaims(token, &ogClaims, keyfunc)

			if err != nil {
				err = errors.Join(ErrParseClaims, err)
			} else {
				apiKey = ogClaims.ApiKey
			}
		}
	} else if hApiKey != "" {
		apiKey = hApiKey
	}

	if apiKey == "" && err == nil {
		err = ErrNoAuth
	}

	return
}
