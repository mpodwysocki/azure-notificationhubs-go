package aznotificationhubs

import (
	"bytes"
	"encoding/json"
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
	Location      string
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

func (n *NotificationHubClient) SendScheduledNotificationWithTags(notificationRequest *NotificationRequest, tags []string, scheduledTime time.Time) (*NotificationResponse, error) {
	tagExpression := strings.Join(tags, "||")
	return n.SendScheduledNotificationWithTagExpression(notificationRequest, tagExpression, scheduledTime)
}

func (n *NotificationHubClient) SendScheduledNotificationWithTagExpression(notificationRequest *NotificationRequest, tagExpression string, scheduledTime time.Time) (*NotificationResponse, error) {
	var correlationId, trackingId, location string
	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)

	requestUri := fmt.Sprintf("%v%v/schedulednotifications/?api-version=%v", fixedHost, n.HubName, apiVersion)

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

	req.Header.Add("Content-Type", notificationRequest.ContentType)
	req.Header.Add("Authorization", sasToken)
	req.Header.Add("ServiceBusNotification-Format", notificationRequest.Platform)
	req.Header.Set("User-Agent", generateUserAgent())
	req.Header.Add("ServiceBusNotification-Tags", tagExpression)
	req.Header.Add("ServiceBusNotification-ScheduleTime", scheduledTime.Format(time.RFC3339))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 201 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", res.StatusCode)
	}

	correlationId = res.Header.Get("x-ms-correlation-request-id")
	trackingId = res.Header.Get("TrackingId")
	location = res.Header.Get("Location")

	return &NotificationResponse{
		CorrelationId: correlationId,
		TrackingId:    trackingId,
		Location:      location,
	}, nil
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
	var correlationId, trackingId, location string
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

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 201 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", res.StatusCode)
	}

	correlationId = res.Header.Get("x-ms-correlation-request-id")
	trackingId = res.Header.Get("TrackingId")
	location = res.Header.Get("Location")

	return &NotificationResponse{
		CorrelationId: correlationId,
		TrackingId:    trackingId,
		Location:      location,
	}, nil
}

func (n *NotificationHubClient) GetInstallation(installationId string) (*Installation, error) {
	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)
	requestUri := fmt.Sprintf("%v%v/installations/%v?api-version=%v", fixedHost, n.HubName, installationId, apiVersion)

	client := &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest(http.MethodGet, requestUri, nil)
	if err != nil {
		return nil, err
	}

	sasToken := n.TokenProvider.GenerateSasToken(n.HostName)

	req.Header.Add("Content-Type", "application/json-patch+json")
	req.Header.Add("Authorization", sasToken)
	req.Header.Set("User-Agent", generateUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", res.StatusCode)
	}

	defer res.Body.Close()

	var installation Installation
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&installation)
	if err != nil {
		return nil, err
	}

	return &installation, nil
}

func (n *NotificationHubClient) CreateOrUpdateInstallation(installation *Installation) (*InstallationResponse, error) {
	var contentLocation string

	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)
	requestUri := fmt.Sprintf("%v%v/installations/%v?api-version=%v", fixedHost, n.HubName, installation.InstallationId, apiVersion)

	installationJSON, err := json.Marshal(installation)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest(http.MethodPut, requestUri, bytes.NewBuffer(installationJSON))
	if err != nil {
		return nil, err
	}

	sasToken := n.TokenProvider.GenerateSasToken(n.HostName)

	req.Header.Add("Content-Type", "application/json-patch+json")
	req.Header.Add("Authorization", sasToken)
	req.Header.Set("User-Agent", generateUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", res.StatusCode)
	}

	contentLocation = res.Header.Get("Content-Location")

	return &InstallationResponse{
		ContentLocation: contentLocation,
	}, nil
}

func (n *NotificationHubClient) PatchInstallation(installationId string, patches []*InstallationPatch) (*InstallationResponse, error) {
	var contentLocation string

	fixedHost := strings.Replace(n.HostName, "sb://", "https://", -1)
	requestUri := fmt.Sprintf("%v%v/installations/%v?api-version=%v", fixedHost, n.HubName, installationId, apiVersion)

	patchesJSON, err := json.Marshal(patches)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 15}
	req, err := http.NewRequest(http.MethodPatch, requestUri, bytes.NewBuffer(patchesJSON))
	if err != nil {
		return nil, err
	}

	sasToken := n.TokenProvider.GenerateSasToken(n.HostName)

	req.Header.Add("Content-Type", "application/json-patch+json")
	req.Header.Add("Authorization", sasToken)
	req.Header.Set("User-Agent", generateUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 204 {
		return nil, fmt.Errorf("invalid response from Azure Notification Hubs: %v", res.StatusCode)
	}

	contentLocation = res.Header.Get("Content-Location")

	return &InstallationResponse{
		ContentLocation: contentLocation,
	}, nil
}
