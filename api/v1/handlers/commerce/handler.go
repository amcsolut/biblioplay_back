package commerce

import (
	"net/http"

	commercedto "api-backend-infinitrum/api/v1/dto/commerce"
	"api-backend-infinitrum/api/v1/middleware"
	commercesvc "api-backend-infinitrum/api/v1/services/commerce"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *commercesvc.Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: commercesvc.NewService(db)}
}

// @Summary Registrar compra (itens pagos)
// @Tags commerce
// @Accept json
// @Produce json
// @Param request body commercedto.CreatePurchaseRequest true "Linhas da compra (obras ou coleções)"
// @Success 201 {object} commercedto.PurchaseResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/purchases [post]
func (h *Handler) CreatePurchase(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req commercedto.CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreatePaidPurchase(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}
