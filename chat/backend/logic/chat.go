package logic

import (
	"encoding/json"
	"os"
)

type FAQQuestion struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type FAQCategory struct {
	Title     string        `json:"title"`
	Questions []FAQQuestion `json:"questions"`
}

type FAQ struct {
	Categories                []FAQCategory `json:"categories"`
	Title                     string        `json:"chat_title"`
	Subtitle                  string        `json:"chat_subtitle"`
	Fallback                  string        `json:"fallback_message"`
	SendBtn                   string        `json:"send_button"`
	LoadingMessage            string        `json:"loading_message"`
	AIErrorMessage            string        `json:"ai_error_message"`
	NotFound                  string        `json:"not_found"`
	Email                     string        `json:"contact_email"`
	Search                    string        `json:"search_placeholder"`
	All                       string        `json:"all_categories"`
	Questions                 string        `json:"questions_title"`
	StartHint                 string        `json:"start_hint"`
	NoResults                 string        `json:"no_results"`
	Reset                     string        `json:"reset_chat"`
	EmailLabel                string        `json:"email_label"`
	NeedAttendant             string        `json:"need_attendant"`
	AIChatButton              string        `json:"ai_chat_button"`
	AIInputLockedHint         string        `json:"ai_input_locked_hint"`
	AIUnlockedMessage         string        `json:"ai_unlocked_message"`
	SatisfactionButton        string        `json:"satisfaction_button"`
	EscalateButton            string        `json:"escalate_button"`
	EscalationCreating        string        `json:"escalation_creating"`
	EscalationCreated         string        `json:"escalation_created"`
	EscalationError           string        `json:"escalation_error"`
	SatisfactionThanks        string        `json:"satisfaction_thanks"`
	InputLockedHint           string        `json:"input_locked_hint"`
	AttendantUnlockedMessage  string        `json:"attendant_unlocked_message"`
	AttendantEscalationNotice string        `json:"attendant_escalation_notice"`
}

func GetFAQ(lang string) (*FAQ, error) {
	data, err := os.ReadFile("./locales/" + lang + "/support.json")
	if err != nil {
		return nil, err
	}

	var faq FAQ
	err = json.Unmarshal(data, &faq)
	if err != nil {
		return nil, err
	}

	return &faq, nil
}

// Future: func AskChat(query string, lang string) string {} for AI
// func LogChatMessage(userID, message string) {}
