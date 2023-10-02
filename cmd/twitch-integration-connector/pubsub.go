package main

import (
	"context"
	"encoding/json"
	"strings"

	"log/slog"

	"slices"

	"github.com/kirides/twitch-integration/twitch"
	"github.com/kirides/twitch-integration/twitch/pubsub"
)

type Redemption struct {
	Title    string `json:"title"`
	Redeemer string `json:"redeemer"`
	Channel  string `json:"channel"`
}

func handlePubSub(ctx context.Context, logger *slog.Logger, cnf twitchCnf, broker eventPublisher) {
	logger = logger.With(logKeyCategory, "pubsub")

	if !cnf.ChannelPointsIntegration {
		logger.Info("integration disabled by configuration.")
		return
	}

	if cnf.OAuthToken == "" {
		return
	}

	logger.Info("Starting Twitch Channelpoints integration.")

	resp, err := twitch.OAuth2Validate(context.Background(), cnf.OAuthToken)
	if err != nil {
		logger.Error("Could not validate OAuth Token", slog.Any("err", err))
		return
	}

	found := slices.Contains(resp.Scopes, "channel:read:redemptions")
	if !found {
		logger.Error("OAuth token does not contain required scope(s)", slog.String("err", "scope not satisfied"), slog.String("scope", "channel:read:redemptions"))
		return
	}

	conn, err := pubsub.Connect(ctx, logger)
	if err != nil {
		logger.Error("failed to connect to twitch pubsub.", slog.Any("err", err))
		return
	}
	defer conn.Close()
	conn.Handle("reward-redeemed", pubsub.H(func(rr *pubsub.RewardRedeemed) {
		logger.Info("Reward redeemed", slog.Any("Reward", rr))

		user := rr.Redemption.User.Login
		user = strings.Replace(user, " ", "", -1)

		redemption := rr.Redemption.Reward.Title
		data, err := json.Marshal(EventEnvelop{Type: "redemption", Data: Redemption{Title: redemption, Redeemer: user, Channel: "-"}})
		if err != nil {
			logger.Error("could not serialize redemption", slog.Any("err", err), slog.String("redeeming_user", rr.Redemption.User.Login))
			return
		}
		logger.Debug("Reward redeemed", slog.String("redeeming_user", rr.Redemption.User.Login), slog.String("reward", redemption))
		broker.Publish(data)
	}))

	go func() {
		if err := conn.ProcessEvents(ctx); err != nil {
			logger.Error("failed to process events.", slog.Any("err", err))
			return
		}
	}()

	if err := conn.Sub(ctx, cnf.OAuthToken, pubsub.TopicChannelPoints); err != nil {
		logger.Error("failed to subscribe to channelpoints topic.", slog.Any("err", err))
	} else {
		logger.Info("subscribed to channelpoints")
	}
	<-ctx.Done()
}
