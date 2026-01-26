package deposit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAddress(c *gin.Context) {
	userID := c.GetString("userID")
	assetID := c.Query("asset")
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset query param required"})
		return
	}

	addr, err := h.service.GenerateAddress(userID, assetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"address": addr, "asset": assetID})
}

// SimulateWebhookRequest mimics a blockchain listener payload
type SimulateWebhookRequest struct {
	TxHash  string  `json:"tx_hash" binding:"required"`
	AssetID string  `json:"asset_id" binding:"required"`
	Address string  `json:"address" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

func (h *Handler) SimulateWebhook(c *gin.Context) {
	var req SimulateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Record incoming deposit
	err := h.service.SimulateWebhook(c.Request.Context(), req.TxHash, req.AssetID, req.Address, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Auto-confirm for simulation purposes (in real life, this is a separate background worker)
	err = h.service.ConfirmDeposit(c.Request.Context(), req.TxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to confirm: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "tx": req.TxHash})
}


