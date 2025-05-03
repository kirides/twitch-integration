package eventsub

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RawEventSubMessage struct {
	Metadata struct {
		MessageID           string    `json:"message_id"`
		MessageType         string    `json:"message_type"`
		MessageTimestamp    time.Time `json:"message_timestamp"`
		SubscriptionType    string    `json:"subscription_type"`
		SubscriptionVersion string    `json:"subscription_version"`
	} `json:"metadata"`
	Payload json.RawMessage `json:"payload"`
}

type RawSubscriptionPayload struct {
	Subscription Subscription           `json:"subscription"`
	Event        json.RawMessage        `json:"event"`
	Condition    map[string]interface{} `json:"condition"`
	Transport    json.RawMessage        `json:"transport"`
}

type AnonymousNotification struct {
	Subscription Subscription    `json:"subscription"`
	Event        json.RawMessage `json:"event"`
}

type RewardUpdateNotification struct {
	Subscription Subscription `json:"subscription"`
	Event        RewardUpdate `json:"event"`
}
type RewardAddNotification struct {
	Subscription Subscription `json:"subscription"`
	Event        RewardAdd    `json:"event"`
}
type ChannelFollowNotification struct {
	Subscription Subscription  `json:"subscription"`
	Event        ChannelFollow `json:"event"`
}
type ChannelCheerNotification struct {
	Subscription Subscription `json:"subscription"`
	Event        ChannelCheer `json:"event"`
}

type Condition struct {
	BroadcasterUserID     string `json:"broadcaster_user_id,omitempty"`
	ModeratorUserId       string `json:"moderator_user_id,omitempty"`
	BroadcasterID         string `json:"broadcaster_id,omitempty"`
	UserID                string `json:"user_id,omitempty"`
	FromBroadcasterUserID string `json:"from_broadcaster_user_id,omitempty"`
	ToBroadcasterUserID   string `json:"to_broadcaster_user_id,omitempty"`
	RewardID              string `json:"reward_id,omitempty"`
	ClientID              string `json:"client_id,omitempty"`
	ConduitID             string `json:"conduit_id,omitempty"`
	CategoryID            string `json:"category_id,omitempty"`
	CampaignID            string `json:"campaign_id,omitempty"`
}

func appendCondition(sb *strings.Builder, value string) {
	if sb.Len() > 0 {
		sb.WriteString(", ")
	}
	sb.WriteString(value)
}

func (c Condition) String() string {
	sb := strings.Builder{}
	if c.BroadcasterUserID != "" {
		appendCondition(&sb, fmt.Sprintf("broadcaster_user_id: %q", c.BroadcasterUserID))
	}
	return sb.String()
}

type Subscription struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Version   string          `json:"version"`
	Status    string          `json:"status"`
	Cost      int             `json:"cost"`
	Condition Condition       `json:"condition"`
	Transport json.RawMessage `json:"transport"`
	CreatedAt time.Time       `json:"created_at"`
}

type RewardUpdate struct {
	ID                   string `json:"id"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	UserInput            string `json:"user_input"`
	Status               string `json:"status"`
	Reward               struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Cost   int    `json:"cost"`
		Prompt string `json:"prompt"`
	} `json:"reward"`
	RedeemedAt time.Time `json:"redeemed_at"`
}

type ChannelFollow struct {
	UserID               string    `json:"user_id"`
	UserLogin            string    `json:"user_login"`
	UserName             string    `json:"user_name"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	FollowedAt           time.Time `json:"followed_at"`
}

type RewardAdd struct {
	ID                   string `json:"id"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	UserInput            string `json:"user_input"`
	Status               string `json:"status"`
	Reward               struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Cost   int    `json:"cost"`
		Prompt string `json:"prompt"`
	} `json:"reward"`
	RedeemedAt time.Time `json:"redeemed_at"`
}

type ChannelCheer struct {
	// when true UserId, UserLogin and UserName are nil
	IsAnonymous          bool    `json:"is_anonymous"`
	UserID               *string `json:"user_id"`
	UserLogin            *string `json:"user_login"`
	UserName             *string `json:"user_name"`
	BroadcasterUserID    string  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string  `json:"broadcaster_user_login"`
	BroadcasterUserName  string  `json:"broadcaster_user_name"`
	Message              string  `json:"message"`
	Bits                 int64   `json:"bits"`
}

type EventSubWelcome struct {
	Session struct {
		ID                      string    `json:"id"`
		Status                  string    `json:"status"`
		ConnectedAt             time.Time `json:"connected_at"`
		KeepaliveTimeoutSeconds int       `json:"keepalive_timeout_seconds"`
		ReconnectURL            *string   `json:"reconnect_url"`
	} `json:"session"`
}

type EventReconnect struct {
	Session struct {
		ID                      string    `json:"id"`
		Status                  string    `json:"status"`
		KeepaliveTimeoutSeconds any       `json:"keepalive_timeout_seconds"`
		ReconnectURL            string    `json:"reconnect_url"`
		ConnectedAt             time.Time `json:"connected_at"`
	} `json:"session"`
}
