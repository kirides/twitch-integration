package pubsub

import "time"

type RewardRedeemed struct {
	Timestamp  time.Time `json:"timestamp"`
	Redemption struct {
		ID   string `json:"id"`
		User struct {
			ID          string `json:"id"`
			Login       string `json:"login"`
			DisplayName string `json:"display_name"`
		} `json:"user"`
		ChannelID  string    `json:"channel_id"`
		RedeemedAt time.Time `json:"redeemed_at"`
		Reward     struct {
			ID                  string `json:"id"`
			ChannelID           string `json:"channel_id"`
			Title               string `json:"title"`
			Prompt              string `json:"prompt"`
			Cost                int    `json:"cost"`
			IsUserInputRequired bool   `json:"is_user_input_required"`
			IsSubOnly           bool   `json:"is_sub_only"`
			Image               struct {
				URL1X string `json:"url_1x"`
				URL2X string `json:"url_2x"`
				URL4X string `json:"url_4x"`
			} `json:"image"`
			DefaultImage struct {
				URL1X string `json:"url_1x"`
				URL2X string `json:"url_2x"`
				URL4X string `json:"url_4x"`
			} `json:"default_image"`
			BackgroundColor string `json:"background_color"`
			IsEnabled       bool   `json:"is_enabled"`
			IsPaused        bool   `json:"is_paused"`
			IsInStock       bool   `json:"is_in_stock"`
			MaxPerStream    struct {
				IsEnabled    bool `json:"is_enabled"`
				MaxPerStream int  `json:"max_per_stream"`
			} `json:"max_per_stream"`
			ShouldRedemptionsSkipRequestQueue bool `json:"should_redemptions_skip_request_queue"`
		} `json:"reward"`
		UserInput string `json:"user_input"`
		Status    string `json:"status"`
	} `json:"redemption"`
}

type BitsEvent struct {
	UserName         string    `json:"user_name"`
	ChannelName      string    `json:"channel_name"`
	UserID           string    `json:"user_id"`
	ChannelID        string    `json:"channel_id"`
	Time             time.Time `json:"time"`
	ChatMessage      string    `json:"chat_message"`
	BitsUsed         int       `json:"bits_used"`
	TotalBitsUsed    int       `json:"total_bits_used"`
	IsAnonymous      bool      `json:"is_anonymous"`
	Context          string    `json:"context"`
	BadgeEntitlement any       `json:"badge_entitlement"`
}
