package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TicketSupportStatusPending = "pending_human_support"
	TicketSupportOriginAIChat  = "ai_chat"
)

type TicketSupport struct {
	ID                      primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID                  primitive.ObjectID     `bson:"userId" json:"userId"`
	UserName                string                 `bson:"userName,omitempty" json:"userName,omitempty"`
	UserEmail               string                 `bson:"userEmail,omitempty" json:"userEmail,omitempty"`
	Status                  string                 `bson:"status" json:"status"`
	Origin                  string                 `bson:"origin" json:"origin"`
	SatisfactionPercent     int                    `bson:"satisfactionPercent" json:"satisfactionPercent"`
	HumanSupportNeedPercent int                    `bson:"humanSupportNeedPercent" json:"humanSupportNeedPercent"`
	Reason                  string                 `bson:"reason,omitempty" json:"reason,omitempty"`
	Messages                []TicketSupportMessage `bson:"messages" json:"messages"`
	CreatedAt               time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt               time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type TicketSupportMessage struct {
	Role      string    `bson:"role" json:"role"`
	Content   string    `bson:"content" json:"content"`
	CreatedAt time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
}
