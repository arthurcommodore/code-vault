package logic

import (
	"strings"
	"time"

	"github.com/arthurcommodore/cotarpreco/internal/model"
	"github.com/arthurcommodore/cotarpreco/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const supportTicketMessageMaxText = 2000

type HumanSupportTicketInput struct {
	Messages          []ChatHistoryMessage
	SupportValidation ChatSatisfactionValidation
}

func CreateHumanSupportTicket(user model.User, input HumanSupportTicketInput) (model.TicketSupport, error) {
	now := time.Now()

	ticket := model.TicketSupport{
		UserID:                  user.ID,
		UserName:                user.Name,
		UserEmail:               user.Email,
		Status:                  model.TicketSupportStatusPending,
		Origin:                  model.TicketSupportOriginAIChat,
		SatisfactionPercent:     input.SupportValidation.SatisfactionPercent,
		HumanSupportNeedPercent: input.SupportValidation.HumanSupportNeedPercent,
		Reason:                  strings.TrimSpace(input.SupportValidation.Reason),
		Messages:                supportTicketMessages(input.Messages),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	insertResult, err := repository.TicketSupportRepo.Insert(ticket)
	if err != nil {
		return model.TicketSupport{}, err
	}

	if objectID, ok := insertResult.InsertedID.(primitive.ObjectID); ok {
		ticket.ID = objectID
	}

	return ticket, nil
}

func supportTicketMessages(messages []ChatHistoryMessage) []model.TicketSupportMessage {
	if len(messages) > chatHistoryMaxItems {
		messages = messages[len(messages)-chatHistoryMaxItems:]
	}

	result := make([]model.TicketSupportMessage, 0, len(messages))
	for _, message := range messages {
		role := strings.TrimSpace(strings.ToLower(message.Type))
		if role == "bot" {
			role = "assistant"
		}
		if role != "user" && role != "assistant" {
			continue
		}

		content := strings.TrimSpace(message.Text)
		if content == "" {
			continue
		}
		contentRunes := []rune(content)
		if len(contentRunes) > supportTicketMessageMaxText {
			content = string(contentRunes[:supportTicketMessageMaxText])
		}

		result = append(result, model.TicketSupportMessage{
			Role:      role,
			Content:   content,
			CreatedAt: message.CreatedAt,
		})
	}

	return result
}
