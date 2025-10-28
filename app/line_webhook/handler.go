package linewebhook

import (
	"errors"
	"net/http"
	"nomnomhub/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"go.uber.org/zap"
)

type handler struct {
	log *zap.Logger
	cfg config.LineWebhook
	bot *messaging_api.MessagingApiAPI
}

func NewHandler(log *zap.Logger, cfg config.LineWebhook) *handler {
	bot, err := messaging_api.NewMessagingApiAPI(cfg.ChannelToken)
	if err != nil {
		log.Fatal("line bot start error", zap.Error(err))
	}
	return &handler{log: log, cfg: cfg, bot: bot}
}

func (h *handler) LineWebhook(c *gin.Context) {
	cb, err := webhook.ParseRequest(h.cfg.ChannelSecret, c.Request)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			h.log.Error("invalid signature")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
			return
		}
		h.log.Error("cannot parse request", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	for _, event := range cb.Events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				if _, err = h.bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							messaging_api.TextMessage{
								Text: message.Text,
							},
							messaging_api.TextMessageV2{
								Text: "Hello! {smile}",
								Substitution: map[string]messaging_api.SubstitutionObjectInterface{
									"smile": &messaging_api.EmojiSubstitutionObject{
										ProductId: "5ac1bfd5040ab15980c9b435",
										EmojiId:   "002",
									},
								},
							},
						},
					},
				); err != nil {
					h.log.Error("cannot reply message", zap.Error(err))
					c.Status(http.StatusInternalServerError)
					return
				} else {
					h.log.Info("sent text reply")
				}
			}
		}

	}

	c.Status(http.StatusOK)
}
