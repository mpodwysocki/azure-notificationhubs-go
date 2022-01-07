package aznotificationhubs

import (
	"strings"
	"testing"
)

const (
	validConnectionString = "<Some-Connection-String>"
	deviceToken           = "<Some-Token>"
	hubName               = "<Some-Hub>"
	messageBody           = `{"aps": { "alert": { "title": "My title", "body": "My body" } } }`
)

func TestParseConnectionString(t *testing.T) {
	parsedConnection, err := FromConnectionString(validConnectionString)
	if parsedConnection == nil || err != nil {
		t.Fatalf(`FromConnectionString = %q, %v`, parsedConnection, err)
	}

	if !strings.EqualFold(parsedConnection.Endpoint, "sb://sdk-sample-namespace.servicebus.windows.net/") {
		t.Fatalf(`ParsedConnection.EndPoint = %q`, parsedConnection.Endpoint)
	}

	if !strings.EqualFold(parsedConnection.KeyName, "NewFullAccessPolicy") {
		t.Fatalf(`ParsedConnection.KeyName = %q`, parsedConnection.KeyName)
	}
}

func TestDirectSend(t *testing.T) {
	client, err := NewNotificationHubClientWithConnectionString(validConnectionString, hubName)
	if client == nil || err != nil {
		t.Fatalf(`NewNotificationHubClientWithConnectionString %v`, err)
	}

	headers := make(map[string]string)
	headers["apns-topic"] = "com.microsoft.XamarinPushTest"
	headers["apns-priority"] = "10"
	headers["apns-push-type"] = "alert"

	contentType := "application/json;charset=utf-8"
	platform := "apple"

	request := &NotificationRequest{
		Message:     messageBody,
		Headers:     headers,
		Platform:    platform,
		ContentType: contentType,
	}

	response, err := client.SendDirectNotification(request, deviceToken)
	if response == nil || err != nil {
		t.Fatalf(`NewNotificationHubClientWithConnectionString %v`, err)
	}
}
