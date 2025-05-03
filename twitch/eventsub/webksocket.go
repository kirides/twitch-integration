package eventsub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/kirides/twitch-integration/twitch"
)

const endpoint = "wss://eventsub.wss.twitch.tv/ws"

type WebsocketConnection struct {
	handler       Handler
	conn          *websocket.Conn
	httpClient    *http.Client
	HTTPHeader    http.Header
	logger        *slog.Logger
	lastKeepalive time.Time
	clientID      string

	OnEvent func(RawEventSubMessage)

	onSubscribe  func(map[string]Condition)
	userToken    string
	wg           sync.WaitGroup
	cancelReadFn func()
	runCtx       context.Context
	readCtx      context.Context
	keepaliveCh  chan time.Time

	session struct {
		ID string
	}
}

func (c *WebsocketConnection) Close() error {
	if c.cancelReadFn != nil {
		c.cancelReadFn()
	}
	c.wg.Wait()
	if c.conn != nil {
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			c.logger.Warn("failed to close old connection.", slog.Any("err", err))
		}
	}
	c.readCtx, c.cancelReadFn = context.WithCancel(context.Background())
	return nil
}

func NewWebsocket(
	clientID string,
	logger *slog.Logger,
	token string,
	handler Handler,
	onSubscribe func(map[string]Condition),
) (*WebsocketConnection, error) {
	conn := &WebsocketConnection{
		clientID:    clientID,
		onSubscribe: onSubscribe,
		logger:      logger,
		userToken:   token,
		handler:     handler,
		httpClient:  http.DefaultClient,
		HTTPHeader:  make(http.Header),
		keepaliveCh: make(chan time.Time),
		runCtx:      context.Background(),
		readCtx:     context.Background(),
	}
	return conn, nil
}

func (c *WebsocketConnection) RunContext(ctx context.Context) error {
	c.runCtx = ctx
	c.Close()

	if err := c.reconnect(ctx, endpoint); err != nil {
		return err
	}
	c.readCtx, c.cancelReadFn = context.WithCancel(ctx)

	c.run(c.readCtx)

	return nil
}

func (c *WebsocketConnection) reconnect(ctx context.Context, url string) error {
	c.Close()

	conn, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPClient: c.httpClient,
		HTTPHeader: c.HTTPHeader,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to %q. %w", url, err)
	}
	var attrs []slog.Attr
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err == nil {
			attrs = append(attrs, slog.String("response", string(data)))
		}
	}
	c.logger.LogAttrs(ctx, slog.LevelInfo, "(Re-)Connected", attrs...)

	c.conn = conn
	return nil
}
func (c *WebsocketConnection) doSubscribe(ctx context.Context) error {
	subscriptions := make(map[string]Condition)
	c.onSubscribe(subscriptions)

	for k, v := range subscriptions {
		version := eventVersions[k]
		if err := c.Subscribe(ctx, SubscriptionInfo{
			Type:      k,
			Version:   version,
			Condition: v,
			Transport: Transport{
				Method:    "websocket",
				SessionID: c.session.ID,
			},
		}); err != nil {
			return fmt.Errorf("failed to subscribe to %q. %w", k, err)
		}
	}
	return nil
}

func (c *WebsocketConnection) run(ctx context.Context) error {
	defer c.Close()

	retry := true
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for ctx.Done() == nil {
			select {
			case <-c.keepaliveCh:
				// OK
				ticker.Reset(time.Minute)
			case <-ticker.C:
				if err := c.reconnect(c.runCtx, endpoint); err != nil {
					c.logger.Error("Failed to connect to twitch.", slog.String("reconnect_url", endpoint), slog.Any("err", err))
					retry = false
					return
				}
				if err := c.doSubscribe(c.runCtx); err != nil {
					c.logger.Error("Failed to subscribe to events",
						slog.Any("err", err),
					)
					retry = false
					return
				}
			}
		}
	}()

	for ctx.Err() == nil {
		t, d, err := c.conn.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.logger.Error("Failed to read message", slog.Any("err", err))
			if retry {
				c.logger.Debug("retry later")
				continue
			}
			return err
		}
		if t != websocket.MessageText {
			c.logger.Warn("Unsupported messagetype", slog.String("messageType", t.String()))
			continue
		}
		var msg RawEventSubMessage
		if err := json.Unmarshal(d, &msg); err != nil {
			c.logger.Warn("could not unmarshal message", slog.Any("err", err))
			continue
		}
		switch msg.Metadata.MessageType {
		case "session_welcome":
			var evt EventSubWelcome
			if err := json.Unmarshal(msg.Payload, &evt); err != nil {
				c.logger.Warn("could not unmarshal message", slog.Any("err", err))
				continue
			}
			c.session.ID = evt.Session.ID
			c.logger.Info("Welcome",
				slog.String("session.id", evt.Session.ID),
				slog.String("session.status", evt.Session.Status),
				slog.String("session.connected_at", evt.Session.ConnectedAt.Format(time.RFC3339)),
				slog.Int("session.keepalive_timeout_seconds", evt.Session.KeepaliveTimeoutSeconds),
				slog.Any("session.reconnect_url", evt.Session.ReconnectURL),
			)
			if err := c.doSubscribe(ctx); err != nil {
				c.logger.Error("Failed to subscribe to events",
					slog.Any("err", err),
				)
				return err
			}
		case "session_keepalive":
			c.lastKeepalive = time.Now()
			c.logger.Debug("Keepalive received")
		case "session_reconnect":
			var evt EventReconnect
			if err := json.Unmarshal(msg.Payload, &evt); err != nil {
				c.logger.Warn("could not unmarshal message. Stopping EventSub.", slog.Any("err", err))
				return err
			}
			if err := c.reconnect(ctx, evt.Session.ReconnectURL); err != nil {
				c.logger.Error("Failed to connect to twitch. Stopping EventSub.", slog.String("reconnect_url", evt.Session.ReconnectURL), slog.Any("err", err))
				return err
			}
			if err := c.doSubscribe(ctx); err != nil {
				c.logger.Error("Failed to subscribe to events",
					slog.Any("err", err),
				)
				return err
			}
		case "notification":
			var evt RawSubscriptionPayload
			if err := json.Unmarshal(msg.Payload, &evt); err != nil {
				c.logger.Warn("could not unmarshal payload", slog.Any("err", err))
				continue
			}
			c.handler.Delegate(evt.Subscription.Type, d)
		}
	}
	return nil
}

func (c *WebsocketConnection) NewAuthdRequest(method, uri string, body io.Reader) (*http.Request, error) {
	// if err := m.UpdateAccessToken(); err != nil {
	// 	return nil, err
	// }

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.userToken))
	return req, nil
}

func (c *WebsocketConnection) DeleteSubscription(id string) error {
	values := url.Values{}
	values.Set("id", id)
	req, err := c.NewAuthdRequest(http.MethodDelete, twitch.QuerySubscriptionsURL(values), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("response error: %s (%d)", resp.Status, resp.StatusCode)
	}
	return nil
}

func (c *WebsocketConnection) Subscribe(ctx context.Context, info SubscriptionInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	c.logger.Debug("Subscribing", slog.String("type", info.Type), slog.String("payload", string(data)))
	body := bytes.NewReader(data)
	req, err := c.NewAuthdRequest(http.MethodPost, twitch.EventSubSubscriptionsURL, body)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, err := io.ReadAll(resp.Body)
		responseText := ""
		if err != nil {
			responseText = string(data)
		}
		return fmt.Errorf("response error: %s (%d): %s", resp.Status, resp.StatusCode, responseText)
	}

	respRdr := io.LimitReader(resp.Body, 4*1024*1024)
	respBody, err := io.ReadAll(respRdr)
	if err != nil {
		return err
	}
	var respData SubscriptionResponse
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return err
	}

	if len(respData.Data) == 0 || len(respData.Data) > 1 {
		return fmt.Errorf("too much response data")
	}
	return nil
}
