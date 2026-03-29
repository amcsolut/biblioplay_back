package community

import (
	"errors"
	"net/http"

	communityDTO "api-backend-infinitrum/api/v1/dto/community"
	"api-backend-infinitrum/api/v1/middleware"
	communitysvc "api-backend-infinitrum/api/v1/services/community"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *communitysvc.Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: communitysvc.NewService(db)}
}

// @Summary Criar comunidade
// @Description Cria a comunidade do autor autenticado (máximo uma por usuário)
// @Tags communities
// @Accept json
// @Produce json
// @Param request body communityDTO.CreateCommunityRequest true "Dados da comunidade"
// @Success 201 {object} communityDTO.CommunityResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities [post]
func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req communityDTO.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.Create(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Listar minhas comunidades
// @Description Lista comunidades em que o usuário autenticado é dono
// @Tags communities
// @Produce json
// @Success 200 {array} communityDTO.CommunityResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities [get]
func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	list, err := h.svc.ListByOwner(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Obter comunidade por ID
// @Description Dono vê sempre; demais usuários só se a comunidade for pública e ativa
// @Tags communities
// @Produce json
// @Param id path string true "ID da comunidade"
// @Success 200 {object} communityDTO.CommunityResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("id")
	res, err := h.svc.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comunidade não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar comunidade
// @Description Apenas o dono pode atualizar
// @Tags communities
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param request body communityDTO.UpdateCommunityRequest true "Campos a atualizar"
// @Success 200 {object} communityDTO.CommunityResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("id")
	var req communityDTO.UpdateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.Update(id, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comunidade não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir comunidade
// @Description Apenas o dono pode excluir
// @Tags communities
// @Param id path string true "ID da comunidade"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("id")
	if err := h.svc.Delete(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comunidade não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
