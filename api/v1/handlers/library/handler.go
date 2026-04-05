package library

import (
	"errors"
	"net/http"

	librarydto "api-backend-infinitrum/api/v1/dto/library"
	"api-backend-infinitrum/api/v1/middleware"
	libsvc "api-backend-infinitrum/api/v1/services/library"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *libsvc.Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: libsvc.NewService(db)}
}

// @Summary Listar minha biblioteca
// @Description Itens adquiridos ou concedidos ao usuário autenticado
// @Tags library
// @Produce json
// @Success 200 {array} librarydto.LibraryItemResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/me/library [get]
func (h *Handler) ListMyLibrary(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	list, err := h.svc.ListLibrary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Adicionar obra ou coleção gratuita à biblioteca
// @Tags library
// @Accept json
// @Produce json
// @Param request body librarydto.AddLibraryItemRequest true "Tipo e ID do item"
// @Success 201 {object} librarydto.LibraryItemResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/me/library [post]
func (h *Handler) AddFreeToLibrary(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req librarydto.AddLibraryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.AddFreeItem(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Remover item gratuito da biblioteca
// @Description Apenas entradas gratuitas adicionadas pelo próprio usuário
// @Tags library
// @Param itemType path string true "book ou collection"
// @Param itemId path string true "ID do item"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/me/library/{itemType}/{itemId} [delete]
func (h *Handler) RemoveFromLibrary(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	itemType := c.Param("itemType")
	itemID := c.Param("itemId")
	if err := h.svc.RemoveFreeItem(userID, itemType, itemID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item não encontrado ou não pode ser removido (apenas entradas gratuitas adicionadas por você)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Conceder acesso a leitor (autor)
// @Description Apenas usuários com papel autor
// @Tags library
// @Accept json
// @Produce json
// @Param request body librarydto.AuthorGrantRequest true "Leitor e item"
// @Success 201 {object} librarydto.LibraryItemResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/authors/grants [post]
func (h *Handler) AuthorGrant(c *gin.Context) {
	authorID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req librarydto.AuthorGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.GrantFromAuthor(authorID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}
