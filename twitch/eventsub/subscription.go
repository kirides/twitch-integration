package eventsub

import (
	"encoding/json"
	"fmt"
)

const (
	SubChannelFollow                                    = "channel.follow"
	SubChannelChannelPointsCustomRewardRedemptionAdd    = "channel.channel_points_custom_reward_redemption.add"
	SubChannelChannelPointsCustomRewardRedemptionUpdate = "channel.channel_points_custom_reward_redemption.update"
	SubChannelCheer                                     = "channel.cheer"
	SubChannelBitsUse                                   = "channel.bits.use"
)

var eventVersions = map[string]string{
	SubChannelChannelPointsCustomRewardRedemptionAdd:    "1",
	SubChannelChannelPointsCustomRewardRedemptionUpdate: "1",
	SubChannelCheer:   "1",
	SubChannelFollow:  "1",
	SubChannelBitsUse: "1",
}

type Handler struct {
	OnChannelFollow                                    func(ChannelFollow)
	OnChannelChannelPointsCustomRewardRedemptionAdd    func(RewardAdd)
	OnChannelChannelPointsCustomRewardRedemptionUpdate func(RewardUpdate)
	OnChannelCheer                                     func(ChannelCheer)
	OnChannelBitsUse                                   func(ChannelBitsUse)
	OnAny                                              func(AnonymousNotification)
}

func (h *Handler) onAny(typed AnonymousNotification) {
	if h.OnAny != nil {
		h.OnAny(typed)
	}
}
func (h *Handler) Delegate(subType string, data []byte) {
	switch subType {
	case SubChannelFollow:
		var typed ChannelFollowNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		if h.OnChannelFollow != nil {
			h.OnChannelFollow(typed.Event)
		}
	case SubChannelChannelPointsCustomRewardRedemptionAdd:
		var typed RewardAddNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		if h.OnChannelChannelPointsCustomRewardRedemptionAdd != nil {
			h.OnChannelChannelPointsCustomRewardRedemptionAdd(typed.Event)
		}
	case SubChannelChannelPointsCustomRewardRedemptionUpdate:
		var typed RewardUpdateNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		if h.OnChannelChannelPointsCustomRewardRedemptionUpdate != nil {
			h.OnChannelChannelPointsCustomRewardRedemptionUpdate(typed.Event)
		}
	case SubChannelCheer:
		var typed ChannelCheerNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		if h.OnChannelCheer != nil {
			h.OnChannelCheer(typed.Event)
		}
	case SubChannelBitsUse:
		var typed ChannelBitsUseNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		if h.OnChannelBitsUse != nil {
			h.OnChannelBitsUse(typed.Event)
		}

	default:
		fmt.Printf("unhandled: %s\n", subType)
		var typed AnonymousNotification
		if err := json.Unmarshal(data, &typed); err != nil {
			fmt.Printf("failed to handle %s\n", subType)
			return
		}
		h.onAny(typed)
	}
}
