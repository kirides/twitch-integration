package twitch

import "net/url"

const (
	IRCWebSocketURL          = "wss://irc-ws.chat.twitch.tv:443"
	PubSubURL                = "wss://pubsub-edge.twitch.tv"
	OAuth2ValidateURL        = "https://id.twitch.tv/oauth2/validate"
	EventSubSubscriptionsURL = "https://api.twitch.tv/helix/eventsub/subscriptions"
	// EventSubSubscriptionsURL  = "http://127.0.0.1:8080/eventsub/subscriptions"

	EventSubURL = "wss://eventsub.wss.twitch.tv/ws"
	// EventSubURL  = "ws://127.0.0.1:8080/ws?keepalive_timeout_seconds=600"
)

func QueryOAuth2TokenURL(values url.Values) string {
	return "https://id.twitch.tv/oauth2/token?" + values.Encode()
}

func QueryUsersURL(values url.Values) string {
	return "https://api.twitch.tv/helix/users?" + values.Encode()
}

func QuerySubscriptionsURL(values url.Values) string {
	return EventSubSubscriptionsURL + "?" + values.Encode()
}
