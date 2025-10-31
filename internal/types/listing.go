package types //nolint:revive

type Listing struct {
	Accepts     []Accept       `json:"accepts"`
	LastUpdated string         `json:"lastUpdated"`
	Metadata    map[string]any `json:"metadata"`
	Resource    string         `json:"resource"`
	Type        string         `json:"type"`
	X402Version int            `json:"x402Version"`
	BazaarExt   map[string]any `json:"bazaarExt,omitempty"`
}

type Accept struct {
	Asset             string        `json:"asset"`
	Description       string        `json:"description,omitempty"`
	Extra             *AcceptExtra  `json:"extra,omitempty"`
	MaxAmountRequired string        `json:"maxAmountRequired"`
	MaxTimeoutSeconds int           `json:"maxTimeoutSeconds"`
	MimeType          string        `json:"mimeType,omitempty"`
	Network           string        `json:"network"`
	OutputSchema      *OutputSchema `json:"outputSchema,omitempty"`
	PayTo             string        `json:"payTo"`
	Resource          string        `json:"resource"`
	Scheme            string        `json:"scheme"`
}

type AcceptExtra struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type OutputSchema struct {
	Input  OutputInput `json:"input"`
	Output any         `json:"output"`
}

type OutputInput struct {
	Method string `json:"method"`
	Type   string `json:"type"`
}
