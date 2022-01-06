package aznotificationhubs

import (
	"strings"
	"testing"
)

const (
	validConnectionString = "Endpoint=sb://some-namespace.servicebus.windows.net/;SharedAccessKeyName=XXXXXXXX;SharedAccessKey=XXXXXXXX"
	deviceToken           = "00fc13adff785122b4ad28809a3420982341241421348097878e577c991de8f0"
	hubName               = "some"
	messageBody           = `{"aps": { "alert": { "title": "My title", "body": "My body" } } }"`
)

func TestParseConnectionString(t *testing.T) {
	parsedConnection, err := FromConnectionString(validConnectionString)
	if parsedConnection == nil || err != nil {
		t.Fatalf(`FromConnectionString = %q, %v`, parsedConnection, err)
	}

	if !strings.EqualFold(parsedConnection.Endpoint, "sb://some-namespace.servicebus.windows.net/") {
		t.Fatalf(`ParsedConnection.EndPoint = %q`, parsedConnection.Endpoint)
	}

	if !strings.EqualFold(parsedConnection.KeyName, "XXXXXXXX") {
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
		t.Fatalf(`client.SendDirectNotification %v`, err)
	}
}
