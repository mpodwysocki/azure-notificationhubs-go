package aznotificationhubs

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	apiVersion = "2015-01"
)

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
	TrackingId    string
	CorrelationId string
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

func generateUserAgent() string {
	return fmt.Sprintf("NHub/%v} (api-origin=GoSDK;)", apiVersion)
}

func (n *NotificationHubClient) SendDirectNotification(notificationRequest *NotificationRequest, deviceToken string) (*NotificationResponse, error) {
	return n.sendNotification(notificationRequest, &deviceToken, nil)
}

func (n *NotificationHubClient) SendNotificationWithTags(notificationRequest *NotificationRequest, tags []string) (*NotificationResponse, error) {
	tagExpression := strings.Join(tags, "||")
	return n.sendNotification(notificationRequest, nil, &tagExpression)
}

func (n *NotificationHubClient) SendNotificationWithTagExpression(notificationRequest *NotificationRequest, tagExpression string) (*NotificationResponse, error) {
	return n.sendNotification(notificationRequest, nil, &tagExpression)
}

func (n *NotificationHubClient) sendNotification(notificationRequest *NotificationRequest, deviceToken *string, tagExpression *string) (*NotificationResponse, error) {
	var correlationId, trackingId string
	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)

	requestUri := fmt.Sprintf("%v%v/messages/?api-version=%v", fixedHost, n.HubName, apiVersion)
	if deviceToken != nil {
		requestUri += "&direct=true"
	}

	messageBody := []byte(notificationRequest.Message)

	client := &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest(http.MethodPost, requestUri, bytes.NewBuffer(messageBody))
	if err != nil {
		return nil, err
	}

	sasToken := n.TokenProvider.GenerateSasToken(n.HostName)

	for headerName, headerValue := range notificationRequest.Headers {
		req.Header.Add(headerName, headerValue)
	}

	if deviceToken != nil {
		req.Header.Add("ServiceBusNotification-DeviceHandle", *deviceToken)
	}

	if tagExpression != nil {
		req.Header.Add("ServiceBusNotification-Tags", *tagExpression)
	}

	req.Header.Add("Content-Type", notificationRequest.ContentType)
	req.Header.Add("Authorization", sasToken)
	req.Header.Add("ServiceBusNotification-Format", notificationRequest.Platform)
	req.Header.Set("User-Agent", generateUserAgent())

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
		CorrelationId: correlationId,
		TrackingId:    trackingId,
	}, nil
}
