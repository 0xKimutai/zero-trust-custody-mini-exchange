package withdrawal

import (
	"mini-exchange/internal/wallet" // For import check
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type WithdrawalRequest struct {
	AssetID   string  `json:"asset_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	ToAddress string  `json:"to_address" binding:"required"`
}

func (h *Handler) RequestWithdrawal(c *gin.Context) {
	userID := c.GetString("userID")
	var req WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.RequestWithdrawal(c.Request.Context(), userID, req.AssetID, req.ToAddress, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "status": "requested"})
}

func (h *Handler) GetHistory(c *gin.Context) {
	userID := c.GetString("userID")
	history, err := h.service.GetHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"withdrawals": history})
}

func (h *Handler) ProcessBatch(c *gin.Context) {
	// Admin endpoint to trigger batch processing
	if err := h.service.ProcessBatch(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "batch_processed"})
}

// Dummy use to prevent unused error if I accidentally keep imports
var _ = wallet.Service{}
var _ = time.Now()
