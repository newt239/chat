package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/link"
)

type LinkHandler struct {
	linkUseCase link.LinkUseCase
}

func NewLinkHandler(linkUseCase link.LinkUseCase) *LinkHandler {
	return &LinkHandler{
		linkUseCase: linkUseCase,
	}
}

// FetchOGP OGP情報取得
func (h *LinkHandler) FetchOGP(c *gin.Context) {
	var input link.FetchOGPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// URLの検証
	if input.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	output, err := h.linkUseCase.FetchOGP(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
