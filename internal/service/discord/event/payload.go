package event

// Payload Docs: https://discord.com/developers/docs/topics/gateway-events#payload-structure
type Payload struct {
	EventName     string                 `json:"t,omitempty"`
	Sequence      int                    `json:"s,omitempty"`
	OperationCode int                    `json:"op"`
	Data          map[string]interface{} `json:"d"`
}
