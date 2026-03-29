package feed

import (
	"errors"
	"net/http"
	"strconv"

	feedDTO "api-backend-infinitrum/api/v1/dto/feed"
	"api-backend-infinitrum/api/v1/middleware"
	feedsvc "api-backend-infinitrum/api/v1/services/feed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *feedsvc.Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: feedsvc.NewService(db)}
}

// @Summary Listar posts do feed da comunidade
// @Description Lista posts publicados (paginação page/limit)
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param page query int false "Página" default(1)
// @Param limit query int false "Itens por página" default(20)
// @Success 200 {array} feedDTO.PostResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts [get]
func (h *Handler) ListPosts(c *gin.Context) {
	communityID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := h.svc.ListPosts(communityID, page, limit)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comunidade não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Criar post no feed
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param request body feedDTO.CreatePostRequest true "Corpo do post"
// @Success 201 {object} feedDTO.PostResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts [post]
func (h *Handler) CreatePost(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	var req feedDTO.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreatePost(communityID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comunidade não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Obter post por ID
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Success 200 {object} feedDTO.PostResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId} [get]
func (h *Handler) GetPost(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	res, err := h.svc.GetPost(communityID, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar post (apenas autor)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param request body feedDTO.UpdatePostRequest true "Campos a atualizar"
// @Success 200 {object} feedDTO.PostResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId} [put]
func (h *Handler) UpdatePost(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	var req feedDTO.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdatePost(communityID, postID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir post (soft delete, apenas autor)
// @Tags feed
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Success 204 "Sem corpo"
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	if err := h.svc.DeletePost(communityID, postID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Media ---

// @Summary Listar mídias do post
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Success 200 {array} feedDTO.PostMediaResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/media [get]
func (h *Handler) ListPostMedia(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	list, err := h.svc.ListPostMedia(communityID, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Adicionar mídia ao post (apenas autor do post)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param request body feedDTO.CreatePostMediaRequest true "Dados da mídia"
// @Success 201 {object} feedDTO.PostMediaResponse
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/media [post]
func (h *Handler) CreatePostMedia(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	var req feedDTO.CreatePostMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreatePostMedia(communityID, postID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Obter mídia por ID
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param mediaId path string true "ID da mídia"
// @Success 200 {object} feedDTO.PostMediaResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/media/{mediaId} [get]
func (h *Handler) GetPostMedia(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	mediaID := c.Param("mediaId")
	res, err := h.svc.GetPostMedia(communityID, postID, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "mídia não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar mídia (apenas autor do post)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param mediaId path string true "ID da mídia"
// @Param request body feedDTO.UpdatePostMediaRequest true "Campos a atualizar"
// @Success 200 {object} feedDTO.PostMediaResponse
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/media/{mediaId} [put]
func (h *Handler) UpdatePostMedia(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	mediaID := c.Param("mediaId")
	var req feedDTO.UpdatePostMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdatePostMedia(communityID, postID, mediaID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "mídia não encontrada"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir mídia (apenas autor do post)
// @Tags feed
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param mediaId path string true "ID da mídia"
// @Success 204 "Sem corpo"
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/media/{mediaId} [delete]
func (h *Handler) DeletePostMedia(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	mediaID := c.Param("mediaId")
	if err := h.svc.DeletePostMedia(communityID, postID, mediaID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "mídia não encontrada"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Comments ---

// @Summary Listar comentários do post
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Success 200 {array} feedDTO.CommentResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments [get]
func (h *Handler) ListComments(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	list, err := h.svc.ListComments(communityID, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Criar comentário
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param request body feedDTO.CreateCommentRequest true "Texto do comentário"
// @Success 201 {object} feedDTO.CommentResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments [post]
func (h *Handler) CreateComment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	var req feedDTO.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateComment(communityID, postID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post não encontrado"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Obter comentário por ID
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Success 200 {object} feedDTO.CommentResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId} [get]
func (h *Handler) GetComment(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	res, err := h.svc.GetComment(communityID, postID, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comentário não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar comentário (apenas autor)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Param request body feedDTO.UpdateCommentRequest true "Novo texto"
// @Success 200 {object} feedDTO.CommentResponse
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId} [put]
func (h *Handler) UpdateComment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	var req feedDTO.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateComment(communityID, postID, commentID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comentário não encontrado"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir comentário (apenas autor)
// @Tags feed
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Success 204 "Sem corpo"
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId} [delete]
func (h *Handler) DeleteComment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	if err := h.svc.DeleteComment(communityID, postID, commentID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comentário não encontrado"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Replies ---

// @Summary Listar respostas do comentário
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Success 200 {array} feedDTO.ReplyResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId}/replies [get]
func (h *Handler) ListReplies(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	list, err := h.svc.ListReplies(communityID, postID, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comentário não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Criar resposta (thread opcional via parent_reply_id)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Param request body feedDTO.CreateReplyRequest true "Texto e reply pai opcional"
// @Success 201 {object} feedDTO.ReplyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId}/replies [post]
func (h *Handler) CreateReply(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	var req feedDTO.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateReply(communityID, postID, commentID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comentário não encontrado"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Obter resposta por ID
// @Tags feed
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Param replyId path string true "ID da resposta"
// @Success 200 {object} feedDTO.ReplyResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId}/replies/{replyId} [get]
func (h *Handler) GetReply(c *gin.Context) {
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	replyID := c.Param("replyId")
	res, err := h.svc.GetReply(communityID, postID, commentID, replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "resposta não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar resposta (apenas autor)
// @Tags feed
// @Accept json
// @Produce json
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Param replyId path string true "ID da resposta"
// @Param request body feedDTO.UpdateReplyRequest true "Novo texto"
// @Success 200 {object} feedDTO.ReplyResponse
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId}/replies/{replyId} [put]
func (h *Handler) UpdateReply(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	replyID := c.Param("replyId")
	var req feedDTO.UpdateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateReply(communityID, postID, commentID, replyID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "resposta não encontrada"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir resposta (apenas autor)
// @Tags feed
// @Param id path string true "ID da comunidade"
// @Param postId path string true "ID do post"
// @Param commentId path string true "ID do comentário"
// @Param replyId path string true "ID da resposta"
// @Success 204 "Sem corpo"
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/communities/{id}/posts/{postId}/comments/{commentId}/replies/{replyId} [delete]
func (h *Handler) DeleteReply(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	communityID := c.Param("id")
	postID := c.Param("postId")
	commentID := c.Param("commentId")
	replyID := c.Param("replyId")
	if err := h.svc.DeleteReply(communityID, postID, commentID, replyID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "resposta não encontrada"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Reactions ---

// @Summary Listar reações de um alvo
// @Description Query: target_type (post|comment|reply) e target_id
// @Tags feed
// @Produce json
// @Param target_type query string true "Tipo do alvo" Enums(post, comment, reply)
// @Param target_id query string true "ID do alvo"
// @Success 200 {array} feedDTO.ReactionResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/feed/reactions [get]
func (h *Handler) ListReactions(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	if targetType == "" || targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query target_type e target_id são obrigatórios"})
		return
	}
	list, err := h.svc.ListReactions(targetType, targetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Criar ou atualizar reação do usuário (upsert)
// @Tags feed
// @Accept json
// @Produce json
// @Param request body feedDTO.UpsertReactionRequest true "Alvo e tipo"
// @Success 200 {object} feedDTO.ReactionResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/feed/reactions [post]
func (h *Handler) UpsertReaction(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req feedDTO.UpsertReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpsertReaction(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Remover reação (apenas quem criou)
// @Tags feed
// @Param reactionId path string true "ID da reação"
// @Success 204 "Sem corpo"
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/feed/reactions/{reactionId} [delete]
func (h *Handler) DeleteReaction(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	reactionID := c.Param("reactionId")
	if err := h.svc.DeleteReaction(reactionID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reação não encontrada"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
