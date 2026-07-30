package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arthurcommodore/cotarpreco/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	chatHistoryTTL      = 7 * 24 * time.Hour
	chatHistoryMaxItems = 80
	chatHistoryMaxText  = 4000
)

type ChatHistoryState struct {
	Messages      []ChatHistoryMessage `json:"messages"`
	InputUnlocked bool                 `json:"inputUnlocked"`
	MessageID     int                  `json:"messageId"`
}

type ChatHistoryMessage struct {
	ID                   int                         `json:"id"`
	Text                 string                      `json:"text"`
	Type                 string                      `json:"type"`
	Source               string                      `json:"source,omitempty"`
	SupportValidation    *ChatSatisfactionValidation `json:"supportValidation,omitempty"`
	Feedback             string                      `json:"feedback,omitempty"`
	HumanSupportTicketID string                      `json:"humanSupportTicketId,omitempty"`
	HumanSupportStatus   string                      `json:"humanSupportStatus,omitempty"`
	HumanSupportError    string                      `json:"humanSupportError,omitempty"`
	CreatedAt            time.Time                   `json:"createdAt,omitempty"`
}

func GetChatHistory(userID string) (ChatHistoryState, error) {
	if repository.RedisRepository == nil {
		return ChatHistoryState{}, nil
	}

	ctx, cancel := redisCtx()
	defer cancel()

	val, err := repository.RedisRepository.Get(ctx, chatHistoryKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return ChatHistoryState{}, nil
		}
		return ChatHistoryState{}, err
	}

	var state ChatHistoryState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return ChatHistoryState{}, err
	}

	return sanitizeChatHistoryState(state), nil
}

func SaveChatHistory(userID string, state ChatHistoryState) (ChatHistoryState, error) {
	state = sanitizeChatHistoryState(state)

	if repository.RedisRepository == nil {
		return state, nil
	}

	data, err := json.Marshal(state)
	if err != nil {
		return state, err
	}

	ctx, cancel := redisCtx()
	defer cancel()

	return state, repository.RedisRepository.Set(ctx, chatHistoryKey(userID), data, chatHistoryTTL).Err()
}

func ClearChatHistory(userID string) error {
	if repository.RedisRepository == nil {
		return nil
	}

	ctx, cancel := redisCtx()
	defer cancel()

	return repository.RedisRepository.Del(ctx, chatHistoryKey(userID)).Err()
}

func chatHistoryKey(userID string) string {
	return fmt.Sprintf("chat:history:%s", userID)
}

func sanitizeChatHistoryState(state ChatHistoryState) ChatHistoryState {
	if len(state.Messages) > chatHistoryMaxItems {
		state.Messages = state.Messages[len(state.Messages)-chatHistoryMaxItems:]
	}

	messages := make([]ChatHistoryMessage, 0, len(state.Messages))
	nextID := 0
	for _, msg := range state.Messages {
		msg.Type = strings.TrimSpace(strings.ToLower(msg.Type))
		if msg.Type != "user" && msg.Type != "bot" {
			continue
		}

		msg.Source = strings.TrimSpace(strings.ToLower(msg.Source))
		if msg.Source == "" {
			if msg.Type == "user" {
				msg.Source = "user"
			} else {
				msg.Source = "faq"
			}
		}
		switch msg.Source {
		case "user", "faq", "ai", "system":
		default:
			msg.Source = "system"
		}

		msg.Text = strings.TrimSpace(msg.Text)
		if msg.Text == "" {
			continue
		}

		textRunes := []rune(msg.Text)
		if len(textRunes) > chatHistoryMaxText {
			msg.Text = string(textRunes[:chatHistoryMaxText])
		}

		if msg.ID < 0 {
			msg.ID = nextID
		}
		if msg.ID >= nextID {
			nextID = msg.ID + 1
		}

		msg.Feedback = sanitizeChatHistoryShortText(msg.Feedback, 40)
		msg.HumanSupportTicketID = sanitizeChatHistoryShortText(msg.HumanSupportTicketID, 80)
		msg.HumanSupportStatus = sanitizeChatHistoryShortText(msg.HumanSupportStatus, 240)
		msg.HumanSupportError = sanitizeChatHistoryShortText(msg.HumanSupportError, 240)
		msg.SupportValidation = sanitizeChatHistorySupportValidation(msg.SupportValidation)

		messages = append(messages, msg)
	}

	state.Messages = messages
	if state.MessageID < nextID {
		state.MessageID = nextID
	}

	return state
}

func sanitizeChatHistoryShortText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}

func sanitizeChatHistorySupportValidation(validation *ChatSatisfactionValidation) *ChatSatisfactionValidation {
	if validation == nil {
		return nil
	}

	clean := NormalizeChatSatisfactionValidation(*validation)
	clean.Reason = sanitizeChatHistoryShortText(clean.Reason, 240)
	return &clean
}
