package aznotificationhubs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type TokenProvider struct {
	KeyName  string
	KeyValue string
}

func NewTokenProvider(keyName string, keyValue string) *TokenProvider {
	return &TokenProvider{
		KeyName:  keyName,
		KeyValue: keyValue,
	}
}

func (t *TokenProvider) GenerateSasToken(uri string) string {
	audience := strings.ToLower(uri)
	sts, expiration := createStringToSign(audience)
	sig := t.signString(sts)
	tokenParams := url.Values{
		"sr":  {audience},
		"sig": {sig},
		"se":  {fmt.Sprintf("%d", expiration)},
		"skn": {t.KeyName},
	}

	return fmt.Sprintf("SharedAccessSignature %s", tokenParams.Encode())
}

func createStringToSign(uri string) (signature string, expiration int64) {
	expiry := time.Now().UTC().Unix() + int64(3600)
	return fmt.Sprintf("%s\n%d", url.QueryEscape(uri), expiry), expiry
}

func (t *TokenProvider) signString(str string) string {
	h := hmac.New(sha256.New, []byte(t.KeyValue))
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

const (
	endpointKey            = "Endpoint"
	sharedAccessKeyNameKey = "SharedAccessKeyName"
	sharedAccessKeyKey     = "SharedAccessKey"
)

type ParsedConnection struct {
	Endpoint string
	KeyName  string
	KeyValue string
}

func FromConnectionString(connectionString string) (*ParsedConnection, error) {
	var endpoint, keyName, keyValue string
	splits := strings.Split(connectionString, ";")
	for _, split := range splits {
		keyValuePair := strings.Split(split, "=")
		if len(keyValuePair) < 2 {
			return nil, errors.New("failed parsing connection string due to unmatched key value separated by '='")
		}

		key := keyValuePair[0]
		value := strings.Join(keyValuePair[1:], "=")
		switch {
		case strings.EqualFold(endpointKey, key):
			endpoint = value
		case strings.EqualFold(sharedAccessKeyNameKey, key):
			keyName = value
		case strings.EqualFold(sharedAccessKeyKey, key):
			keyValue = value
		}
	}

	if endpoint == "" {
		return nil, fmt.Errorf("key %q must not be empty", endpointKey)
	}

	if keyName == "" {
		return nil, fmt.Errorf("key %q must not be empty", sharedAccessKeyNameKey)
	}

	if keyValue == "" {
		return nil, fmt.Errorf("key %q must not be empty", sharedAccessKeyKey)
	}

	return &ParsedConnection{
		Endpoint: endpoint,
		KeyName:  keyName,
		KeyValue: keyValue,
	}, nil
}
