package linewebhook

import (
	"fmt"
	"net/http"
	"nomnomhub/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot"
)

type handler struct {
	bot *linebot.Client
}

func NewHandler(cfg config.LineWebhook) *handler {
	bot, err := linebot.New(cfg.ChannelSecret, cfg.ChannelToken)
	if err != nil {
		panic(fmt.Errorf("failed to init linebot: %v", err))
	}
	return &handler{bot: bot}
}

func (h *handler) LineWebhook(c *gin.Context) {
	events, err := h.bot.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				reply := fmt.Sprintf("You said: %s", message.Text)
				if _, err := h.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
					fmt.Println("reply error:", err)
				}
			}
		}
	}

	c.Status(http.StatusOK)
}
