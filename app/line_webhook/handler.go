package linewebhook

import (
	"context"
	"errors"
	"net/http"
	"nomnomhub/app"
	"nomnomhub/internal/config"
	"nomnomhub/internal/model"
	"nomnomhub/internal/opengraph"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type handler struct {
	log     *zap.Logger
	cfg     config.LineWebhook
	bot     *messaging_api.MessagingApiAPI
	storage *app.Storage
}

func NewHandler(log *zap.Logger, cfg config.LineWebhook, storage *app.Storage) *handler {
	bot, err := messaging_api.NewMessagingApiAPI(cfg.ChannelToken)
	if err != nil {
		log.Fatal("cannot start line bot", zap.Error(err))
	}
	return &handler{log: log, cfg: cfg, bot: bot, storage: storage}
}

func (h *handler) Test(c *gin.Context) {
	c.Status(http.StatusOK)
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
			// check if bot is mentioned
			var textMessageContent webhook.TextMessageContent
			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				textMessageContent = message
				if !isMentioned(textMessageContent) {
					continue
				}
			}

			// get user id
			userLineID := ""
			switch s := e.Source.(type) {
			case webhook.UserSource:
				userLineID = s.UserId
			case webhook.RoomSource:
				userLineID = s.UserId
			case webhook.GroupSource:
				userLineID = s.UserId
			default:
				h.log.Error("unknown source")
				c.Status(http.StatusInternalServerError)
				return
			}

			// create user if not exist
			user, err := h.storage.GetUserByLineID(c, userLineID)
			if err != nil {
				h.log.Error("GetUserByLineID", zap.Error(err))
				c.Status(http.StatusInternalServerError)
				return
			}
			if user == nil {
				profile, err := h.bot.GetProfile(userLineID)
				if err != nil {
					h.log.Error("cannot get user line profile", zap.Error(err))
					c.Status(http.StatusInternalServerError)
					return
				}

				if err := h.storage.UpsertUserByLineID(c, model.User{
					LineID:      userLineID,
					DisplayName: profile.DisplayName,
				}); err != nil {
					h.log.Error("UpsertUserByLineID", zap.Error(err))
					c.Status(http.StatusInternalServerError)
					return
				}
			}

			// parse link
			links := extractPossibleLinks(textMessageContent.Text)

			// add places
			places := make([]model.Place, len(links))
			for _, link := range links {
				places = append(places, model.Place{
					ID:   uuid.New(),
					Name: "",
					URL:  link,
					AddedBy: uuid.NullUUID{
						UUID:  user.ID,
						Valid: true,
					},
				})
			}

			if err := h.storage.WithTx(c, func(c context.Context, tx bun.Tx) error {
				if err := h.storage.CreatePlaces(c, places); err != nil {
					return err
				}
				return nil
			}); err != nil {
				h.log.Error("CreatePlaces", zap.Error(err))
				c.Status(http.StatusInternalServerError)
				return
			}

			// ask for tags

			// ask other users for review
			replyMessages := make([]messaging_api.MessageInterface, 0)
			for _, link := range links {
				previewData, err := opengraph.FetchPreview(c, link)
				if err != nil {
					continue
				}
				replyMessages = append(replyMessages, BuildPreviewFlex(previewData))
			}

			if _, err = h.bot.ReplyMessage(
				&messaging_api.ReplyMessageRequest{
					ReplyToken: e.ReplyToken,
					Messages:   replyMessages,
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

	c.Status(http.StatusOK)
}

func isMentioned(msg webhook.TextMessageContent) bool {
	if msg.Mention == nil {
		return false
	}
	for _, m := range msg.Mention.Mentionees {
		switch v := m.(type) {
		case webhook.UserMentionee:
			if v.IsSelf {
				return true
			}
		}
	}
	return false
}

var looseLinkRegex = regexp.MustCompile(`(?i)\b((?:[a-z0-9-]+\.)+[a-z]{2,}(?:/[^\s]*)?)`)

func extractPossibleLinks(text string) []string {
	return looseLinkRegex.FindAllString(text, -1)
}

func BuildPreviewFlex(preview *opengraph.Preview) messaging_api.FlexMessage {
	var hero *messaging_api.FlexImage
	if preview.Image != "" {
		hero = &messaging_api.FlexImage{
			Url:         preview.Image,
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  "cover",
			Action: &messaging_api.UriAction{
				Label: "เปิดลิงก์",
				Uri:   preview.URL,
			},
		}
	}

	bodyContents := []messaging_api.FlexComponentInterface{
		&messaging_api.FlexText{
			Text:   safeText(preview.Title, "🔗 ลิงก์จากครอบครัว"),
			Weight: "bold",
			Size:   "lg",
			Wrap:   true,
		},
		&messaging_api.FlexText{
			Text:  preview.URL,
			Size:  "sm",
			Color: "#AAAAAA",
			Wrap:  true,
		},
	}

	footerButtons := []messaging_api.FlexComponentInterface{
		button("❌ ไม่ผ่าน", "#FF3B30", "ไม่ผ่าน", messaging_api.FlexButtonSTYLE_SECONDARY),
		button("😐 ก็โอเค", "#FF9500", "ก็โอเค", messaging_api.FlexButtonSTYLE_SECONDARY),
		button("🔥 ไป!!", "#34C759", "ไป!!", messaging_api.FlexButtonSTYLE_PRIMARY),
	}

	return messaging_api.FlexMessage{
		AltText: "มีร้านใหม่มาดูหน่อย!",
		Contents: messaging_api.FlexBubble{
			Hero: hero,
			Body: &messaging_api.FlexBox{
				Layout:   "vertical",
				Contents: bodyContents,
			},
			Footer: &messaging_api.FlexBox{
				Layout:   "horizontal",
				Spacing:  "md",
				Contents: footerButtons,
			},
		},
	}
}

func button(label, color, text string, style messaging_api.FlexButtonSTYLE) *messaging_api.FlexButton {
	return &messaging_api.FlexButton{
		Style: style,
		Color: color,
		Action: &messaging_api.MessageAction{
			Label: label,
			Text:  text,
		},
	}
}

func safeText(text, fallback string) string {
	if strings.TrimSpace(text) == "" {
		return fallback
	}
	return text
}
