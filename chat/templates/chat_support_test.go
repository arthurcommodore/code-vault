package components

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestChatSubscriptionRequiredDoesNotInitializeChat(t *testing.T) {
	var buffer bytes.Buffer
	translate := func(key string, _ ...string) string {
		switch key {
		case "chat_subscription_required_title":
			return "Chat disponível para assinantes"
		case "chat_subscription_required_action":
			return "Ver assinatura"
		default:
			return key
		}
	}

	if err := ChatSubscriptionRequired(translate, "pt").Render(context.Background(), &buffer); err != nil {
		t.Fatal(err)
	}

	html := buffer.String()
	if strings.Contains(html, `x-data="chat(`) {
		t.Fatalf("expected locked chat page to avoid chat initialization, got %s", html)
	}
	if !strings.Contains(html, "Chat disponível para assinantes") {
		t.Fatalf("expected subscription required title, got %s", html)
	}
	if !strings.Contains(html, `/pt/app/subscription`) {
		t.Fatalf("expected subscription link, got %s", html)
	}
}
