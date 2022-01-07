# Azure Notification Hubs for Go (Unofficial)

This is the unofficial Azure Notification Hubs SDK for Go.  

## Usage

Below are code snippets for each scenario that the SDK covers.

### Direct Send

This example uses the [Direct Send API](https://docs.microsoft.com/en-us/rest/api/notificationhubs/direct-send) to send a message to an Apple device through APNs.

```go
const (
	validConnectionString = "<Some-Connection-String>"
	deviceToken           = "<Some-Token>"
	hubName               = "<Some-Hub>"
	messageBody           = `{"aps": { "alert": { "title": "My title", "body": "My body" } } }`
)

func TestDirectSend() {
	client, err := NewNotificationHubClientWithConnectionString(validConnectionString, hubName)
	if client == nil || err != nil {
		panic(err)
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
		panic(err)
	}
}
```

## Status

- Added Direct Send

### TODO

- Installation Support
- Registration Support
- Tag-Based Send
- Template Send
- Scheduled Send

## LICENSE

MIT