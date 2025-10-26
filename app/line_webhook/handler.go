package linewebhook

import (
	"net/http"
	"nomnomhub/internal/config"

	"github.com/gin-gonic/gin"
)

type handler struct {
	cfg config.LineWebhook
}

func NewHandler(cfg config.LineWebhook) *handler {
	return &handler{
		cfg: cfg,
	}
}

func (h *handler) LineWebhook(c *gin.Context) {
	c.Status(http.StatusOK)
}
