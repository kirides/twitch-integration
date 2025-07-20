package main

/*
	Testing the Eventsub:

	1. Download twitch developer CLI from here https://github.com/twitchdev/twitch-cli/releases/latest get the one that fits your system, e.g. xxxxx_Windows_x86_64.zip for x64 bit windows.
	2. Open a terminal window in the same directory as `twitch.exe`
	3. Launch the dummy server `./twitch.exe event websocket start-server`
	4. Run twitch-integration-connector so the configuration is created/updated.
	5. End with CTRL+C
	6. Adjust `twitch-integration-connector.json` to point to dummy websocket server from 3.: `"eventsub_url": "ws://127.0.0.1:8080/ws"` and start it again.
	7. Open a NEW terminal window in the same directory as `twitch.exe`
	8. Send some event to the application: `./twitch.exe event trigger --item-name "reward 523523" --transport=websocket channel.channel_points_custom_reward_redemption.add`
*/

import (
	"context"
	"encoding/json"
	"strings"

	"log/slog"

	"slices"

	"github.com/kirides/twitch-integration/twitch"
	"github.com/kirides/twitch-integration/twitch/eventsub"
)

type Redemption struct {
	Title    string `json:"title"`
	Redeemer string `json:"redeemer"`
	Channel  string `json:"channel"`
}

type BitsEvent struct {
	BitsUsed int    `json:"bitsUsed"`
	User     string `json:"user"`
	Channel  string `json:"channel"`
}

func handleEventSub(ctx context.Context, logger *slog.Logger, cnf twitchCnf, broker eventPublisher) {
	logger = logger.With(logKeyCategory, "eventsub")

	if !cnf.ChannelPointsIntegration && !cnf.BitsIntegration {
		logger.Info("integration disabled by configuration.")
		return
	}

	if strings.Contains(cnf.OAuthToken, "id.twitch.tv") || cnf.OAuthToken == "" {
		logger.Info("No credentials. Integration disabled.")
		return
	}

	logger.Info("Starting Twitch Channelpoints integration.")

	resp, err := twitch.OAuth2Validate(context.Background(), cnf.OAuthToken)
	if err != nil {
		logger.Error("Could not validate OAuth Token", slog.Any("err", err))
		return
	}

	subFns := []func(subscriptions map[string]eventsub.Condition){}

	evHandler := &eventsub.Handler{
		OnChannelChannelPointsCustomRewardRedemptionAdd: func(rr eventsub.RewardAdd) {
			logger.Info("Reward redeemed", slog.Any("Reward", rr))

			userWithoutSpaces := rr.UserLogin
			userWithoutSpaces = strings.Replace(userWithoutSpaces, " ", "", -1)

			redemption := rr.Reward.Title
			data, err := json.Marshal(EventEnvelop{Type: "redemption", Data: Redemption{Title: redemption, Redeemer: userWithoutSpaces, Channel: "-"}})
			if err != nil {
				logger.Error("could not serialize redemption", slog.Any("err", err), slog.String("redeeming_user", rr.UserLogin))
				return
			}
			logger.Debug("Reward redeemed", slog.String("redeeming_user", rr.UserLogin), slog.String("reward", redemption))
			broker.Publish(data)
		},
		OnChannelCheer: func(rr eventsub.ChannelCheer) {

			logger.Info("Bits used", slog.Any("BitsEvent", rr))
			username := "anonymous"
			if !rr.IsAnonymous && rr.UserName != nil {
				username = *rr.UserName
			}

			userWithoutSpaces := username
			userWithoutSpaces = strings.Replace(userWithoutSpaces, " ", "", -1)

			data, err := json.Marshal(EventEnvelop{Type: "bits", Data: BitsEvent{BitsUsed: int(rr.Bits), User: userWithoutSpaces, Channel: rr.BroadcasterUserLogin}})
			if err != nil {
				logger.Error("could not serialize bits", slog.Any("err", err), slog.String("user", username))
				return
			}
			logger.Debug("Reward redeemed", slog.String("user", username), slog.Int64("bits", rr.Bits))
			broker.Publish(data)
		},
	}

	if cnf.ChannelPointsIntegration {
		if found := slices.Contains(resp.Scopes, "channel:read:redemptions"); !found {
			logger.Error("OAuth token does not contain required scope(s)", slog.String("err", "scope not satisfied"), slog.String("scope", "channel:read:redemptions"))
			return
		}
		subFns = append(subFns, func(subscriptions map[string]eventsub.Condition) {
			subscriptions[eventsub.SubChannelChannelPointsCustomRewardRedemptionAdd] = eventsub.Condition{
				BroadcasterUserID: resp.UserID,
			}
		})
	}

	if cnf.BitsIntegration {
		if found := slices.Contains(resp.Scopes, "bits:read"); !found {
			logger.Error("OAuth token does not contain required scope(s)", slog.String("err", "scope not satisfied"), slog.String("scope", "bits:read"))
			return
		}

		subFns = append(subFns, func(subscriptions map[string]eventsub.Condition) {
			subscriptions[eventsub.SubChannelCheer] = eventsub.Condition{
				BroadcasterUserID: resp.UserID,
			}
		})
	}

	if len(subFns) == 0 {
		logger.Info("No integrations enabled")
		return
	}

	conn, err := eventsub.NewWebsocket(
		resp.ClientID,
		logger,
		cnf.OAuthToken,
		*evHandler,
		func(m map[string]eventsub.Condition) {
			for _, v := range subFns {
				v(m)
			}
		})

	if err != nil {
		logger.Error("failed to connect to twitch eventsub.", slog.Any("err", err))
		return
	}
	conn.EventSubURL = cnf.EventSubURL

	defer conn.Close()

	conn.OnEvent = func(e eventsub.RawEventSubMessage) {
		logger.Debug("EVENT RECEIVED", slog.String("type", e.Metadata.MessageType), slog.String("data", string(e.Payload)))
	}

	go func() {
		if err := conn.RunContext(ctx); err != nil {
			logger.Error("failed to process events.", slog.Any("err", err))
			return
		}
	}()
	<-ctx.Done()
}
