package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arthurcommodore/cotarpreco/internal/logic/gpt"
)

const chatHumanSupportNeedThreshold = 70

type ChatSatisfactionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatSatisfactionValidation struct {
	SatisfactionPercent     int    `json:"satisfactionPercent"`
	HumanSupportNeedPercent int    `json:"humanSupportNeedPercent"`
	ShowEscalation          bool   `json:"showEscalation"`
	Reason                  string `json:"reason,omitempty"`
}

type ChatSatisfactionReviewInput struct {
	History         []ChatSatisfactionMessage `json:"history"`
	UserMessage     string                    `json:"userMessage"`
	AssistantAnswer string                    `json:"assistantAnswer"`
}

func DefaultChatSatisfactionValidation() ChatSatisfactionValidation {
	return ChatSatisfactionValidation{SatisfactionPercent: 70}
}

func EvaluateChatSatisfactionWithAI(
	ctx context.Context,
	apiKey string,
	model string,
	input ChatSatisfactionReviewInput,
) (ChatSatisfactionValidation, error) {
	input = sanitizeChatSatisfactionReviewInput(input)
	if input.UserMessage == "" || input.AssistantAnswer == "" {
		return DefaultChatSatisfactionValidation(), nil
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return DefaultChatSatisfactionValidation(), err
	}

	resp, err := gpt.CallOpenAIChat(ctx, apiKey, model, []gpt.Message{
		{
			Role: "system",
			Content: `Compare a última mensagem do usuário com a resposta da IA e avalie a satisfação do atendimento.
			Use o histórico só como contexto.
			Responda apenas JSON válido neste formato:
			{"satisfactionPercent":0-100,"humanSupportNeedPercent":0-100,"reason":"curto"}`,
		},
		{Role: "user", Content: string(payload)},
	})
	if err != nil {
		return DefaultChatSatisfactionValidation(), err
	}

	var validation ChatSatisfactionValidation
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &validation); err != nil {
		return DefaultChatSatisfactionValidation(), fmt.Errorf("invalid satisfaction JSON: %w", err)
	}

	return NormalizeChatSatisfactionValidation(validation), nil
}

func NormalizeChatSatisfactionValidation(validation ChatSatisfactionValidation) ChatSatisfactionValidation {
	validation.SatisfactionPercent = clampChatPercent(validation.SatisfactionPercent)
	validation.HumanSupportNeedPercent = clampChatPercent(validation.HumanSupportNeedPercent)
	validation.Reason = sanitizeChatSatisfactionText(validation.Reason, 240)
	validation.ShowEscalation = validation.HumanSupportNeedPercent >= chatHumanSupportNeedThreshold && validation.SatisfactionPercent <= 55
	return validation
}

func sanitizeChatSatisfactionReviewInput(input ChatSatisfactionReviewInput) ChatSatisfactionReviewInput {
	input.UserMessage = sanitizeChatSatisfactionText(input.UserMessage, 1500)
	input.AssistantAnswer = sanitizeChatSatisfactionText(input.AssistantAnswer, 2500)
	if len(input.History) > 8 {
		input.History = input.History[len(input.History)-8:]
	}

	history := make([]ChatSatisfactionMessage, 0, len(input.History))
	for _, item := range input.History {
		role := strings.TrimSpace(strings.ToLower(item.Role))
		if role == "bot" {
			role = "assistant"
		}
		content := sanitizeChatSatisfactionText(item.Content, 1000)
		if (role == "user" || role == "assistant") && content != "" {
			history = append(history, ChatSatisfactionMessage{Role: role, Content: content})
		}
	}
	input.History = history
	return input
}

func sanitizeChatSatisfactionText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > limit {
		return string([]rune(value)[:limit])
	}
	return value
}

func clampChatPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
