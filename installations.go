package aznotificationhubs

type Installation struct {
	InstallationId     string                           `json:"installationId"`
	UserId             string                           `json:"userId,omitempty"`
	LastActiveOn       string                           `json:"lastActiveOn,omitempty"`
	ExpirationTime     string                           `json:"expirationTime,omitempty"`
	LastUpdate         string                           `json:"lastUpdate,omitempty"`
	Platform           string                           `json:"platform"`
	PushChannel        string                           `json:"pushChannel"`
	ExpiredPushChannel bool                             `json:"expiredPushChannel,omitempty"`
	Tags               []string                         `json:"tags,omitempty"`
	Templates          map[string]*InstallationTemplate `json:"templates,omitempty"`
}

type InstallationTemplate struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers,omitempty"`
	Expiry  string            `json:"expiry,omitempty"`
	Tags    []string          `json:"tags,omitempty"`
}

type InstallationPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
}

type InstallationResponse struct {
	ContentLocation string
}
