package aznotificationhubs

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	audience := strings.ToLower(url.QueryEscape(uri))
	sts, expiration := createStringToSign(audience)
	sig := t.signString(sts)
	return fmt.Sprintf("SharedAccessSignature sr=%s&sig=%s&se=%s&skn=%s", audience, sig, expiration, t.KeyName)
}

func createStringToSign(uri string) (signature string, expiration string) {
	expiry := time.Now().UTC().Add(time.Hour).Round(time.Second).Unix()
	expiryTime := strconv.FormatInt(expiry, 10)
	audience := strings.ToLower(url.QueryEscape(uri))
	return audience + "\n" + expiryTime, expiryTime
}

func (t *TokenProvider) signString(str string) string {
	h := hmac.New(sha256.New, []byte(t.KeyValue))
	h.Write([]byte(str))
	encodedSig := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(encodedSig)
}

const (
	endpointKey            = "Endpoint"
	sharedAccessKeyNameKey = "SharedAccessKeyName"
	sharedAccessKeyKey     = "SharedAccessKey"
	apiVersion             = "2015-01"
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

type NotificationHubClient struct {
	HubName       string
	HostName      string
	TokenProvider *TokenProvider
}

type NotificationRequest struct {
	Message     string
	Headers     map[string]string
	Platform    string
	ContentType string
}

type NotificationResponse struct {
	TrackingId     string
	NotificationId string
	CorrelationId  string
}

func NewNotificationHubClientWithConnectionString(connectionString string, hubName string) (*NotificationHubClient, error) {
	parsedConnection, err := FromConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	tokenProvider := NewTokenProvider(parsedConnection.KeyName, parsedConnection.KeyValue)

	return &NotificationHubClient{
		HubName:       hubName,
		HostName:      parsedConnection.Endpoint,
		TokenProvider: tokenProvider,
	}, nil
}

func (n *NotificationHubClient) SendDirectNotification(notificationRequest *NotificationRequest, deviceToken string) (*NotificationResponse, error) {
	var correlationId, trackingId string
	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)
	signatureHost := strings.Replace(n.HostName, "sb://", "http://", -1)

	requestUri := fmt.Sprintf("%v%v/messages/?api-version=%v&direct=true", fixedHost, n.HubName, apiVersion)
	fmt.Printf("URI: %v\n", requestUri)

	messageBody := []byte(notificationRequest.Message)

	client := &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest(http.MethodPost, requestUri, bytes.NewBuffer(messageBody))
	if err != nil {
		return nil, err
	}

	sasToken := n.TokenProvider.GenerateSasToken(signatureHost)
	sasToken = strings.Replace(sasToken, "%3D", "%3d", -1)
	fmt.Printf("Signature Host: %v", signatureHost)
	fmt.Printf("SAS Token: %v", sasToken)

	for headerName, headerValue := range notificationRequest.Headers {
		req.Header.Add(headerName, headerValue)
	}

	req.Header.Add("Content-Type", notificationRequest.ContentType)
	req.Header.Add("Authorization", sasToken)
	req.Header.Add("ServiceBusNotification-DeviceHandle", deviceToken)
	req.Header.Add("ServiceBusNotification-Format", notificationRequest.Platform)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", resp.StatusCode)
	}

	correlationId = resp.Header.Get("x-ms-correlation-request-id")
	trackingId = resp.Header.Get("TrackingId")

	return &NotificationResponse{
		CorrelationId:  correlationId,
		TrackingId:     trackingId,
		NotificationId: "",
	}, nil
}
