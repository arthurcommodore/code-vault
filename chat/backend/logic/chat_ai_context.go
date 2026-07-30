package logic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arthurcommodore/cotarpreco/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	chatAIRecentRecordsLimit = 10
	chatAIContextTextLimit   = 120
)

type chatAIContext struct {
	InterfaceLang string                `json:"interfaceLang"`
	RecentRecords chatAIRecentRecords   `json:"recentRecords"`
	Links         ChatAINavigationLinks `json:"links"`
}

type chatAIRecentRecords struct {
	Products   []chatAIRecentRecord `json:"products"`
	Categories []chatAIRecentRecord `json:"categories"`
	Quotes     []chatAIRecentRecord `json:"quotes"`
}

type chatAIRecentRecord struct {
	Name        string `json:"name"`
	Code        string `json:"code,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type ChatAINavigationLinks struct {
	Products     string `json:"products"`
	Categories   string `json:"categories"`
	Quotes       string `json:"quotes"`
	Subscription string `json:"subscription"`
	Chat         string `json:"chat"`
}

func BuildChatAISystemPrompt(userID, lang, question string) (string, error) {
	faqDomain := buildFAQDomain(lang)

	dataContext, err := buildChatAIContext(userID, lang)
	if err != nil {
		return "", err
	}

	dataJSON, err := json.MarshalIndent(dataContext, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`Você é o assistente de suporte do Cotarpreco.

	Regras:
	- Responda dúvidas sobre uso do Cotarpreco com objetividade.
	- Use o histórico recente da conversa para entender continuidade, reclamações e dúvidas repetidas.
	- Priorize resolver automaticamente antes de atendimento humano.
	- Não incentive atendimento humano por iniciativa própria.
	- Detecte o idioma da última mensagem do usuário e responda nesse idioma. Se estiver ambíguo, use o idioma de interface: %s.
	- Quando fizer sentido, indique a tela correta usando os links do contexto.
	- O contexto de registros recentes é apenas uma amostra dos últimos itens criados ou atualizados. Se algo não estiver nele, oriente o usuário a conferir a tela correta.
	- Não invente valores, documentos, preços, status ou saldos.
	- Não afirme que criou, alterou ou excluiu registros; oriente o usuário a usar a tela correta.
	- Se a pergunta fugir do domínio do Cotarpreco, responda brevemente que você só pode ajudar com dúvidas da plataforma.

	FAQ:
	%s
	Contexto resumido do usuário:
	%s`, chatAILanguageName(lang), faqDomain, string(dataJSON)), nil
}

func buildChatAIContext(userID, lang string) (chatAIContext, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		LogSystem("Error primitive.ObjectIDFromHex, func BuildChatAIContext", err, userID)
		return chatAIContext{}, err
	}

	recentRecords, err := chatAIRecentRecordsForUser(userObjectID)
	if err != nil {
		return chatAIContext{}, err
	}

	return chatAIContext{
		InterfaceLang: chatAISafeLang(lang),
		RecentRecords: recentRecords,
		Links:         chatAINavigationLinksForLang(lang),
	}, nil
}

func chatAIRecentRecordsForUser(userObjectID primitive.ObjectID) (chatAIRecentRecords, error) {
	products, err := repository.ProductRepository.FindLatest(bson.M{"userId": userObjectID}, chatAIRecentRecordsLimit)
	if err != nil {
		return chatAIRecentRecords{}, err
	}

	categories, err := repository.CategoryRepo.FindLatest(bson.M{"userId": userObjectID}, chatAIRecentRecordsLimit)
	if err != nil {
		return chatAIRecentRecords{}, err
	}

	quotes, err := repository.QuoteRepo.FindLatest(bson.M{"userId": userObjectID}, chatAIRecentRecordsLimit)
	if err != nil {
		return chatAIRecentRecords{}, err
	}

	records := chatAIRecentRecords{
		Products:   make([]chatAIRecentRecord, 0, len(products)),
		Categories: make([]chatAIRecentRecord, 0, len(categories)),
		Quotes:     make([]chatAIRecentRecord, 0, len(quotes)),
	}

	for _, product := range products {
		records.Products = append(records.Products, chatAIRecentRecord{
			Name:      chatAIShortText(product.Name, chatAIContextTextLimit),
			Code:      chatAIShortText(product.Code, chatAIContextTextLimit),
			Unit:      chatAIShortText(product.Unit, 40),
			UpdatedAt: chatAIFormatDate(product.UpdatedAt),
		})
	}

	for _, category := range categories {
		records.Categories = append(records.Categories, chatAIRecentRecord{
			Name:        chatAIShortText(category.Name, chatAIContextTextLimit),
			Description: chatAIShortText(category.Description, chatAIContextTextLimit),
			UpdatedAt:   chatAIFormatDate(category.UpdatedAt),
		})
	}

	for _, quote := range quotes {
		records.Quotes = append(records.Quotes, chatAIRecentRecord{
			Name:        chatAIShortText(quote.Title, chatAIContextTextLimit),
			Description: chatAIShortText(quote.Notes, chatAIContextTextLimit),
			UpdatedAt:   chatAIFormatDate(quote.UpdatedAt),
		})
	}

	return records, nil
}

func chatAIShortText(value string, limit int) string {
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

func chatAIFormatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func buildFAQDomain(lang string) string {
	faq, err := GetFAQ(lang)
	if err != nil && lang != "pt" {
		faq, err = GetFAQ("pt")
	}
	if err != nil || faq == nil {
		return "FAQ indisponível."
	}

	var b strings.Builder
	for _, category := range faq.Categories {
		b.WriteString("- ")
		b.WriteString(category.Title)
		b.WriteString(":\n")
		for _, question := range category.Questions {
			b.WriteString("  Pergunta: ")
			b.WriteString(question.Q)
			b.WriteString("\n  Resposta: ")
			b.WriteString(question.A)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func chatAINavigationLinksForLang(lang string) ChatAINavigationLinks {
	lang = chatAISafeLang(lang)

	return ChatAINavigationLinks{
		Products:     "/" + lang + "/app/products",
		Categories:   "/" + lang + "/app/categories",
		Quotes:       "/" + lang + "/app/quotes",
		Subscription: "/" + lang + "/app/subscription",
		Chat:         "/" + lang + "/app/chat",
	}
}

func chatAISafeLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "pt", "en", "es":
		return strings.ToLower(strings.TrimSpace(lang))
	default:
		return "pt"
	}
}

func chatAILanguageName(lang string) string {
	switch chatAISafeLang(lang) {
	case "en":
		return "English"
	case "es":
		return "Español"
	default:
		return "Português do Brasil"
	}
}
