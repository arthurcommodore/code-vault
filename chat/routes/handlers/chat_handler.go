package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arthurcommodore/cotarpreco/internal/logic"
	"github.com/arthurcommodore/cotarpreco/internal/logic/gpt"
	"github.com/arthurcommodore/cotarpreco/internal/model"
	"github.com/arthurcommodore/cotarpreco/internal/utils"
	"github.com/gin-gonic/gin"
)

type chatAIRequest struct {
	Message string                 `json:"message" validate:"required,max=1500"`
	History []chatAIHistoryMessage `json:"history"`
}

type chatHumanSupportRequest struct {
	Messages          []logic.ChatHistoryMessage       `json:"messages" validate:"required,min=1,max=80"`
	SupportValidation logic.ChatSatisfactionValidation `json:"supportValidation"`
}

type chatAIHistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var fieldErrorsChatAI = map[string]string{
	"Message": "message_invalid_data",
}

const chatAIMessageMaxText = 1500

func requireChatAccess(c *gin.Context, t func(string, ...string) string, logContext string) (model.User, bool) {
	user, err := logic.GetUserByCookie(c)
	if err != nil {
		logic.LogSystem("Error logic.GetUserByCookie, "+logContext, err)
		utils.JSONError(c, http.StatusInternalServerError,
			map[string]any{"message": t("message_internal_error")},
		)
		return user, false
	}

	user, err = logic.IsChatAccessAllowed(user)
	if err != nil {
		if errors.Is(err, logic.ErrChatSubscriptionRequired) {
			utils.JSONError(c, http.StatusForbidden,
				map[string]any{"message": t("message_chat_subscription_required")},
			)
			return user, false
		}

		logic.LogSystem("Error logic.IsChatAccessAllowed, "+logContext, err, user.ID.Hex())
		utils.JSONError(c, http.StatusInternalServerError,
			map[string]any{"message": t("message_internal_error")},
		)
		return user, false
	}

	return user, true
}

func ChatIA() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := utils.AutoTranslator(c)

		user, ok := requireChatAccess(c, t, "func ChatIA")
		if !ok {
			return
		}

		var req chatAIRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logic.LogSystem("Error c.ShouldBindJSON(&req), func ChatIA", err)
			utils.JSONError(c, http.StatusBadRequest, gin.H{
				"message": t("message_invalid_data"),
			})
			return
		}

		req.Message = strings.TrimSpace(req.Message)
		req.Message = truncateRunes(req.Message, chatAIMessageMaxText)
		if err := validate.Struct(req); err != nil {
			utils.JSONError(c, utils.StatusBadRequest, gin.H{
				"message": t(logic.GetFriendlyError(err, fieldErrorsChatAI)),
			})
			return
		}

		userID := user.ID.Hex()
		if !allowHandlerAction(c, "chat-ai-user", userID, 30, time.Hour, t) {
			return
		}

		apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		if apiKey == "" {
			logic.LogSystem("OPENAI_API_KEY is empty, func ChatIA", nil, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
		if model == "" {
			model = "gpt-4o-mini"
		}

		lang := chatRequestLang(c, user)
		systemPrompt, err := logic.BuildChatAISystemPrompt(userID, lang, req.Message)
		if err != nil {
			logic.LogSystem("Error logic.BuildChatAISystemPrompt, func ChatIA", err, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		messages := []gpt.Message{
			{Role: "system", Content: systemPrompt},
		}
		messages = append(messages, sanitizeChatHistory(req.History)...)
		messages = append(messages, gpt.Message{Role: "user", Content: req.Message})

		ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
		defer cancel()

		resp, err := gpt.CallOpenAIChat(ctx, apiKey, model, messages)
		if err != nil {
			logic.LogSystem("Error gpt.CallOpenAIChat, func ChatIA", err, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		answer := strings.TrimSpace(resp.Choices[0].Message.Content)
		if answer == "" {
			answer = t("message_chat_ai_empty_answer")
		}

		supportValidation := logic.DefaultChatSatisfactionValidation()
		reviewCtx, reviewCancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer reviewCancel()
		supportValidation, err = logic.EvaluateChatSatisfactionWithAI(reviewCtx, apiKey, model, logic.ChatSatisfactionReviewInput{
			History:         chatSatisfactionMessages(req.History),
			UserMessage:     req.Message,
			AssistantAnswer: answer,
		})
		if err != nil {
			logic.LogSystem("Error logic.EvaluateChatSatisfactionWithAI, func ChatIA", err, userID)
			supportValidation = logic.DefaultChatSatisfactionValidation()
		}

		utils.JSON(c, http.StatusOK, gin.H{
			"message":           t("message_chat_ai_answered"),
			"answer":            answer,
			"supportValidation": supportValidation,
		})
	}
}

func sanitizeChatHistory(history []chatAIHistoryMessage) []gpt.Message {
	if len(history) > 10 {
		history = history[len(history)-10:]
	}

	messages := make([]gpt.Message, 0, len(history))
	for _, item := range history {
		role := strings.TrimSpace(strings.ToLower(item.Role))
		if role != "user" && role != "assistant" {
			continue
		}

		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if len([]rune(content)) > 1000 {
			content = string([]rune(content)[:1000])
		}

		messages = append(messages, gpt.Message{
			Role:    role,
			Content: content,
		})
	}

	return messages
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func chatSatisfactionMessages(history []chatAIHistoryMessage) []logic.ChatSatisfactionMessage {
	if len(history) > 12 {
		history = history[len(history)-12:]
	}

	messages := make([]logic.ChatSatisfactionMessage, 0, len(history))
	for _, item := range history {
		role := strings.TrimSpace(strings.ToLower(item.Role))
		if role != "user" && role != "assistant" {
			continue
		}

		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}

		messages = append(messages, logic.ChatSatisfactionMessage{
			Role:    role,
			Content: content,
		})
	}

	return messages
}

func chatRequestLang(c *gin.Context, user model.User) string {
	if lang, err := c.Cookie("lang"); err == nil && strings.TrimSpace(lang) != "" {
		if normalized := utils.NormalizeLang(lang); normalized != "" {
			return normalized
		}
	}

	if strings.TrimSpace(user.Lang) != "" {
		if normalized := utils.NormalizeLang(user.Lang); normalized != "" {
			return normalized
		}
	}

	return utils.DetectBrowserLangByHeader(c.GetHeader("Accept-Language"))
}

func ChatHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := utils.AutoTranslator(c)

		user, ok := requireChatAccess(c, t, "func ChatHistory")
		if !ok {
			return
		}
		userID := user.ID.Hex()

		state, err := logic.GetChatHistory(userID)
		if err != nil {
			logic.LogSystem("Error logic.GetChatHistory, func ChatHistory", err, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		utils.JSON(c, http.StatusOK, state)
	}
}

func SaveChatHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := utils.AutoTranslator(c)

		user, ok := requireChatAccess(c, t, "func SaveChatHistory")
		if !ok {
			return
		}

		var state logic.ChatHistoryState
		if err := c.ShouldBindJSON(&state); err != nil {
			logic.LogSystem("Error c.ShouldBindJSON(&state), func SaveChatHistory", err)
			utils.JSON(c, http.StatusBadRequest, gin.H{
				"message": t("message_invalid_data"),
			})
			return
		}

		userID := user.ID.Hex()

		state, err := logic.SaveChatHistory(userID, state)
		if err != nil {
			logic.LogSystem("Error logic.SaveChatHistory, func SaveChatHistory", err, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		utils.JSON(c, http.StatusOK, state)
	}
}

func ClearChatHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := utils.AutoTranslator(c)

		user, ok := requireChatAccess(c, t, "func ClearChatHistory")
		if !ok {
			return
		}
		userID := user.ID.Hex()

		if err := logic.ClearChatHistory(userID); err != nil {
			logic.LogSystem("Error logic.ClearChatHistory, func ClearChatHistory", err, userID)
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		utils.JSON(c, http.StatusOK, gin.H{
			"message": t("message_chat_history_cleared"),
		})
	}
}

func CreateChatHumanSupportTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := utils.AutoTranslator(c)

		user, ok := requireChatAccess(c, t, "func CreateChatHumanSupportTicket")
		if !ok {
			return
		}

		var req chatHumanSupportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logic.LogSystem("Error c.ShouldBindJSON(&req), func CreateChatHumanSupportTicket", err)
			utils.JSON(c, http.StatusBadRequest, gin.H{
				"message": t("message_invalid_data"),
			})
			return
		}

		if err := validate.Struct(req); err != nil {
			utils.JSONError(c, utils.StatusBadRequest, gin.H{
				"message": t("message_invalid_data"),
			})
			return
		}

		if !allowHandlerAction(c, "chat-ticket-user", user.ID.Hex(), 10, time.Hour, t) {
			return
		}

		ticket, err := logic.CreateHumanSupportTicket(user, logic.HumanSupportTicketInput{
			Messages:          req.Messages,
			SupportValidation: logic.NormalizeChatSatisfactionValidation(req.SupportValidation),
		})
		if err != nil {
			logic.LogSystem("Error logic.CreateHumanSupportTicket, func CreateChatHumanSupportTicket", err, user.ID.Hex())
			utils.JSONError(c, http.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		utils.JSON(c, http.StatusCreated, gin.H{
			"message": t("message_human_support_ticket_created"),
			"ticket":  ticket,
		})
	}
}
