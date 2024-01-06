package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Microsoft/go-winio"
	"go.uber.org/zap"
)

const (
	configFileName = "twitch-integration.json"
)

type chatCommand struct {
	Actions     []string `json:"actions"`
	CooldownSec int32    `json:"cooldown_sec"`
	Message     string   `json:"message"`
}

type config struct {
	Debug          bool           `json:"debug"`
	Twitch         twitch         `json:"twitch"`
	StreamElements streamElements `json:"streamElements"`
}
type twitch struct {
	Rewards map[string][]string    `json:"rewards"`
	Chat    map[string]chatCommand `json:"chat"`
}
type streamElements struct {
	Perks map[string][]string `json:"perks"`
}

func defaultConfig() config {
	return config{
		Debug: false,
		StreamElements: streamElements{
			Perks: map[string][]string{
				"Item1": {"TWI_XXX", "TWI_YYY"},
			},
		},
		Twitch: twitch{
			Rewards: map[string][]string{
				"Item1": {"TWI_XXX", "TWI_YYY"},
			},
			Chat: map[string]chatCommand{
				"#help": {
					Actions:     []string{"XXXXXXXXXXXXXXXXXXXX"},
					Message:     "Dies hier wird im Chat angezeigt",
					CooldownSec: 5,
				},
			},
		},
	}
}

func main() {
	// DLL
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer cancel()
	// cnf, _ := readConfigJson()
	// handleEventPipe(ctx, cnf, log.Default())
}

func handleEventPipe(ctx context.Context, cnf config, logger *zap.Logger) error {
	conn, err := winio.DialPipeAccess(ctx, `\\.\pipe\__TwitchIntegration_Kirides_Conn`, syscall.GENERIC_READ)
	if err != nil {
		return err
	}
	defer conn.Close()
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	logger.Debug("connected to event pipe")

	buffer := make([]byte, 16384)
	for {
		if _, err := io.ReadFull(conn, buffer[:2]); err != nil {
			return fmt.Errorf("could not read message size. %w", err)
		}
		size := binary.LittleEndian.Uint16(buffer[:2])
		if int(size) > len(buffer) {
			return fmt.Errorf("unexpectedly large message")
		}
		if _, err := io.ReadFull(conn, buffer[:size]); err != nil {
			return fmt.Errorf("could not read message. %w", err)
		}
		type envelope struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		var event envelope
		if err := json.Unmarshal(buffer[:size], &event); err != nil {
			return fmt.Errorf("could not deserialize message. %w", err)
		}

		switch event.Type {
		case "chat":
			type ChatMessage struct {
				Text    string `json:"text"`
				Sender  string `json:"sender"`
				Channel string `json:"channel"`
			}
			var chatMessage ChatMessage
			if err := json.Unmarshal(event.Data, &chatMessage); err != nil {
				return fmt.Errorf("could not deserialize chat message event. %w", err)
			}
			if fn, ok := cnf.Twitch.Chat[chatMessage.Text]; ok {
				logger.Info("Event accepted", zap.String("sender", chatMessage.Sender), zap.String("text", chatMessage.Text), zap.Strings("actions", fn.Actions))
				for _, fn := range fn.Actions {
					fn := strings.TrimSpace(fn)
					enqueueEvent(fmt.Sprintf("CHAT %s %s", chatMessage.Sender, fn))
				}
			}
		case "redemption":
			type Redemption struct {
				Title    string `json:"title"`
				Redeemer string `json:"redeemer"`
				Channel  string `json:"channel"`
			}
			var redeption Redemption
			if err := json.Unmarshal(event.Data, &redeption); err != nil {
				return fmt.Errorf("could not deserialize redeption event. %w", err)
			}
			logger.Debug("reward triggered", zap.String("redeemer", redeption.Redeemer), zap.String("reward", redeption.Title))
			if fn, ok := cnf.Twitch.Rewards[strings.ToUpper(redeption.Title)]; ok {
				logger.Info("handling reward", zap.String("redeemer", redeption.Redeemer), zap.String("reward", redeption.Title), zap.Strings("actions", fn))
				for _, fn := range fn {
					fn := strings.TrimSpace(fn)
					enqueueEvent(fmt.Sprintf("REWARD_ADD %s %s", redeption.Redeemer, fn))
				}
			}
		case "streamelements-perk":
			type Redemption struct {
				Title    string `json:"title"`
				Redeemer string `json:"redeemer"`
				Channel  string `json:"channel"`
			}
			var redeption Redemption
			if err := json.Unmarshal(event.Data, &redeption); err != nil {
				return fmt.Errorf("could not deserialize streamelements-perk event. %w", err)
			}
			logger.Debug("StreamElements perk triggered", zap.String("redeemer", redeption.Redeemer), zap.String("perk", redeption.Title))
			if fn, ok := cnf.StreamElements.Perks[strings.ToUpper(redeption.Title)]; ok {
				logger.Info("handling StreamElements perk", zap.String("redeemer", redeption.Redeemer), zap.String("perk", redeption.Title), zap.Strings("actions", fn))
				for _, fn := range fn {
					fn := strings.TrimSpace(fn)
					enqueueEvent(fmt.Sprintf("REWARD_ADD %s %s", redeption.Redeemer, fn))
				}
			}
		}
	}
}

func allKeysToUpper(m map[string][]string) map[string][]string {
	copy := make(map[string][]string, len(m))

	for k, v := range m {
		copy[strings.ToUpper(k)] = v
	}
	return copy
}

func prettyJson(data any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(data)
	return buf.Bytes(), err
}

func readConfigJson() (config, error) {
	cnf := defaultConfig()
	execPath, err := os.Executable()
	if err != nil {
		app.logger.Error("failed to locate current executable", zap.Error(err))
		return cnf, err
	}

	configContent, err := os.ReadFile(filepath.Join(filepath.Dir(execPath), configFileName))
	if err != nil {
		if os.IsNotExist(err) {
			def, err := prettyJson(cnf)
			if err != nil {
				app.logger.Error("failed to marshal default config", zap.Error(err))
				return cnf, err
			}
			os.WriteFile(filepath.Join(filepath.Dir(execPath), configFileName), def, 0640)
		}
	} else {
		if err := json.Unmarshal(configContent, &cnf); err != nil {
			app.logger.Error("failed to unmarshal config", zap.Error(err))
			return cnf, err
		}
	}

	if cnf.Twitch.Rewards == nil {
		cnf.Twitch.Rewards = make(map[string][]string)
	}
	if cnf.StreamElements.Perks == nil {
		cnf.StreamElements.Perks = make(map[string][]string)
	}
	cnf.Twitch.Rewards = allKeysToUpper(cnf.Twitch.Rewards)
	cnf.StreamElements.Perks = allKeysToUpper(cnf.StreamElements.Perks)

	if cnf.Twitch.Chat == nil {
		cnf.Twitch.Chat = make(map[string]chatCommand)
	}
	return cnf, nil
}
