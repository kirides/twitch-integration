package eventsub

import "time"

type SubscriptionInfo struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Condition Condition `json:"condition"`
	Transport Transport `json:"transport"`
	// // An ID that identifies the WebSocket to send notifications to. When you connect to EventSub using WebSockets, the server returns the ID in the Welcome message. Specify this field only if method is set to websocket
	// SessionID string `json:"session_id,omitempty"`
	// // An ID that identifies the conduit to send notifications to. When you create a conduit, the server returns the conduit ID. Specify this field only if method is set to conduit.
	// ConduitID string `json:"conduit_id,omitempty"`
}

type Transport struct {
	Method   string `json:"method"`
	Callback string `json:"callback,omitempty"`
	Secret   string `json:"secret,omitempty"`
	// An ID that identifies the WebSocket that notifications are sent to. Included only if method is set to websocket
	SessionID string `json:"session_id,omitempty"`
}

type SubscriptionResponse struct {
	Data         []Data `json:"data"`
	Total        int    `json:"total"`
	TotalCost    int    `json:"total_cost"`
	MaxTotalCost int    `json:"max_total_cost"`
}

type Data struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Type      string        `json:"type"`
	Version   string        `json:"version"`
	Cost      int           `json:"cost"`
	Condition Condition     `json:"condition"`
	Transport TransportSlim `json:"transport"`
	CreatedAt time.Time     `json:"created_at"`
}

type TransportSlim struct {
	// The transport method. Possible values are
	//
	// - webhook
	//
	// - websocket
	//
	// - conduit
	Method string `json:"method"`
	// The callback URL where the notifications are sent. Included only if method is set to webhook
	Callback string `json:"callback,omitempty"`
	// An ID that identifies the WebSocket that notifications are sent to. Included only if method is set to websocket
	SessionID string `json:"session_id,omitempty"`
}
