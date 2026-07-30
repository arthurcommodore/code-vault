package logic

import "testing"

func TestNormalizeChatSatisfactionValidation(t *testing.T) {
	got := NormalizeChatSatisfactionValidation(ChatSatisfactionValidation{
		SatisfactionPercent:     -10,
		HumanSupportNeedPercent: 120,
		ShowEscalation:          true,
		Reason:                  "  não resolveu  ",
	})

	if got.SatisfactionPercent != 0 {
		t.Fatalf("SatisfactionPercent = %d, want 0", got.SatisfactionPercent)
	}
	if got.HumanSupportNeedPercent != 100 {
		t.Fatalf("HumanSupportNeedPercent = %d, want 100", got.HumanSupportNeedPercent)
	}
	if !got.ShowEscalation {
		t.Fatal("ShowEscalation = false, want true")
	}
	if got.Reason != "não resolveu" {
		t.Fatalf("Reason = %q", got.Reason)
	}
}

func TestNormalizeChatSatisfactionValidationBlocksWeakEscalation(t *testing.T) {
	got := NormalizeChatSatisfactionValidation(ChatSatisfactionValidation{
		SatisfactionPercent:     80,
		HumanSupportNeedPercent: 90,
		ShowEscalation:          true,
	})

	if got.ShowEscalation {
		t.Fatal("ShowEscalation = true, want false for high satisfaction")
	}
}
