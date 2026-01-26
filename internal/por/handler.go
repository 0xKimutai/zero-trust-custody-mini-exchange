package por

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Generate(c *gin.Context) {
	assetID := c.Query("asset")
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset required"})
		return
	}

	root, err := h.service.GenerateSnapshot(c.Request.Context(), assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"asset": assetID, "merkle_root": root})
}

// TODO: Get latest root endpoint
func (h *Handler) GetLatestRoot(c *gin.Context) {
	// Not implemented in Service yet, but placeholder
	c.JSON(http.StatusOK, gin.H{"status": "not implemented yet"})
}
