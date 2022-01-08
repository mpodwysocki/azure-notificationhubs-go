package aznotificationhubs

import (
	"testing"
)

const (
	connectionString = "Endpoint=sb://my-namespace.servicebus.windows.net/;SharedAccessKeyName=key-name;SharedAccessKey=secret"
	deviceToken      = "00fc13adff785122b4ad28809a3420982341241421348097878e577c991de8f0"
	hubName          = "my-hub"
	messageBody      = `{"aps": { "alert": { "title": "My title", "body": "My body" } } }`
)

func TestDirectSend(t *testing.T) {
	client, err := NewNotificationHubClientWithConnectionString(connectionString, hubName)
	if client == nil || err != nil {
		t.Fatalf(`NewNotificationHubClientWithConnectionString %v`, err)
	}

	headers := make(map[string]string)
	headers["apns-topic"] = "com.microsoft.ExampleApp"
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
		t.Fatalf(`SendDirectNotification %v`, err)
	}
}

func TestTagsSend(t *testing.T) {
	client, err := NewNotificationHubClientWithConnectionString(connectionString, hubName)
	if client == nil || err != nil {
		t.Fatalf(`NewNotificationHubClientWithConnectionString %v`, err)
	}

	headers := make(map[string]string)
	headers["apns-topic"] = "com.microsoft.ExampleApp"
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

	tags := []string{"language_en", "country_US"}

	response, err := client.SendNotificationWithTags(request, tags)
	if response == nil || err != nil {
		t.Fatalf(`SendNotificationWithTags %v`, err)
	}
}

func TestTagExpression(t *testing.T) {
	client, err := NewNotificationHubClientWithConnectionString(connectionString, hubName)
	if client == nil || err != nil {
		t.Fatalf(`NewNotificationHubClientWithConnectionString %v`, err)
	}

	headers := make(map[string]string)
	headers["apns-topic"] = "com.microsoft.ExampleApp"
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

	tagExpression := "language_en&&country_US"

	response, err := client.SendNotificationWithTagExpression(request, tagExpression)
	if response == nil || err != nil {
		t.Fatalf(`SendNotificationWithTagExpression %v`, err)
	}
}
