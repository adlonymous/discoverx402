package types

type Listing struct {
	Accepts     []Accept        `json:"accepts"`
	LastUpdated string          `json:"lastUpdated"`       // RFC3339 string
	Metadata    map[string]any  `json:"metadata"`          // can be empty object
	Resource    string          `json:"resource"`          // main resource URL
	Type        string          `json:"type"`              // e.g., "http"
	X402Version int             `json:"x402Version"`       // e.g., 1
	BazaarExt   map[string]any  `json:"bazaarExt,omitempty"` // your optional extensions
}


type Accept struct {
	Asset             string        `json:"asset"`                       // token address or identifier
	Description       string        `json:"description,omitempty"`       // optional
	Extra             *AcceptExtra  `json:"extra,omitempty"`             // optional
	MaxAmountRequired string        `json:"maxAmountRequired"`           // atomic units
	MaxTimeoutSeconds int           `json:"maxTimeoutSeconds"`           // seconds
	MimeType          string        `json:"mimeType,omitempty"`          // expected response MIME type
	Network           string        `json:"network"`                     // e.g., "base"
	OutputSchema      *OutputSchema `json:"outputSchema,omitempty"`      // optional
	PayTo             string        `json:"payTo"`                       // payment address
	Resource          string        `json:"resource"`                    // actual paid endpoint (duplicate of top-level OK)
	Scheme            string        `json:"scheme"`                      // e.g., "exact"
}

type AcceptExtra struct {
	Name    string `json:"name,omitempty"`    // e.g., "USD Coin"
	Version string `json:"version,omitempty"` // e.g., "2"
}

type OutputSchema struct {
	Input  OutputInput `json:"input"`
	Output any         `json:"output"` // null or a schema object
}

type OutputInput struct {
	Method string `json:"method"` // "GET", "POST", etc.
	Type   string `json:"type"`   // typically "http"
}
