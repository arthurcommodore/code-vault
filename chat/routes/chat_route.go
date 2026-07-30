package routes

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/arthurcommodore/cotarpreco/internal/logic"
	"github.com/arthurcommodore/cotarpreco/internal/model"
	"github.com/arthurcommodore/cotarpreco/internal/routes/handlers"
	components "github.com/arthurcommodore/cotarpreco/internal/template"
	"github.com/arthurcommodore/cotarpreco/internal/utils"
	"github.com/gin-gonic/gin"
)

func ChatPages(r *gin.RouterGroup) {
	utils.HandleTrailingSlash(r, r.GET, "/chat", func(c *gin.Context) {
		t := c.MustGet("t").(func(string, ...string) string)
		lang := c.Param("lang")

		user, err := logic.GetUserByCookie(c)
		if err != nil {
			logic.LogSystem("Error logic.GetUserByCookie, func ChatPages", err)
			utils.JSONError(c,
				utils.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
			return
		}

		comp, user, err := chatComponentForUser(user, t, lang)
		if err != nil {
			if errors.Is(err, logic.ErrChatSubscriptionRequired) {
				c.Status(http.StatusForbidden)
			} else {
				logic.LogSystem("Error chatComponentForUser, func ChatPages", err, user.ID.Hex())
				utils.JSONError(c,
					utils.StatusInternalServerError,
					map[string]any{"message": t("message_internal_error")},
				)
				return
			}
		}

		err = comp.Render(c.Request.Context(), c.Writer)
		if err != nil {
			logic.LogSystem("Erro components.ChatSupport, func ChatPages", err)
			utils.JSONError(c,
				utils.StatusInternalServerError,
				map[string]any{"message": t("message_internal_error")},
			)
		}
	})
}

func chatComponentForUser(user model.User, t func(string, ...string) string, lang string) (templ.Component, model.User, error) {
	user, err := logic.IsChatAccessAllowed(user)
	if err != nil {
		if errors.Is(err, logic.ErrChatSubscriptionRequired) {
			return components.ChatSubscriptionRequired(t, lang), user, err
		}
		return nil, user, err
	}

	return components.ChatSupport(t, lang), user, nil
}

func ChatActions(r *gin.RouterGroup) {
	utils.HandleTrailingSlash(r, r.GET, "/chat/history", handlers.ChatHistory())
	utils.HandleTrailingSlash(r, r.PUT, "/chat/history", handlers.SaveChatHistory())
	utils.HandleTrailingSlash(r, r.DELETE, "/chat/history", handlers.ClearChatHistory())
	utils.HandleTrailingSlash(r, r.POST, "/chat/ia", handlers.ChatIA())
	utils.HandleTrailingSlash(r, r.POST, "/chat/human-support", handlers.CreateChatHumanSupportTicket())
}
