package logic

import (
	"strings"
	"testing"
)

func TestSupportTicketMessagesTruncateContent(t *testing.T) {
	oversized := strings.Repeat("a", supportTicketMessageMaxText+10)

	got := supportTicketMessages([]ChatHistoryMessage{
		{Type: "user", Text: oversized},
	})

	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if len([]rune(got[0].Content)) != supportTicketMessageMaxText {
		t.Fatalf("content length = %d, want %d", len([]rune(got[0].Content)), supportTicketMessageMaxText)
	}
}
