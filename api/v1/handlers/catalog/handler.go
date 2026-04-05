package catalog

import (
	"errors"
	"net/http"

	catalogDTO "api-backend-infinitrum/api/v1/dto/catalog"
	"api-backend-infinitrum/api/v1/middleware"
	catalogsvc "api-backend-infinitrum/api/v1/services/catalog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *catalogsvc.Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{svc: catalogsvc.NewService(db)}
}

// --- Books ---

// @Summary Criar obra no catálogo
// @Description Cria livro para o autor autenticado
// @Tags catalog
// @Accept json
// @Produce json
// @Param request body catalogDTO.CreateBookRequest true "Dados da obra"
// @Success 201 {object} catalogDTO.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books [post]
func (h *Handler) CreateBook(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req catalogDTO.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateBook(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Listar obras do autor
// @Description Lista todas as obras do usuário autenticado
// @Tags catalog
// @Produce json
// @Success 200 {array} catalogDTO.BookResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books [get]
func (h *Handler) ListBooks(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	list, err := h.svc.ListBooks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Obter obra por ID
// @Description Apenas o autor da obra
// @Tags catalog
// @Produce json
// @Param bookId path string true "ID da obra"
// @Success 200 {object} catalogDTO.BookResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId} [get]
func (h *Handler) GetBook(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("bookId")
	res, err := h.svc.GetBook(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar obra
// @Description Apenas o autor da obra
// @Tags catalog
// @Accept json
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param request body catalogDTO.UpdateBookRequest true "Campos a atualizar"
// @Success 200 {object} catalogDTO.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId} [put]
func (h *Handler) UpdateBook(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("bookId")
	var req catalogDTO.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateBook(id, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir obra
// @Description Apenas o autor; remove também capítulos em cascata
// @Tags catalog
// @Param bookId path string true "ID da obra"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId} [delete]
func (h *Handler) DeleteBook(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("bookId")
	if err := h.svc.DeleteBook(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Ebook chapters (rotas aninhadas: /catalog/books/:bookId/ebook-chapters) ---

// @Summary Listar capítulos ebook
// @Tags catalog
// @Produce json
// @Param bookId path string true "ID da obra"
// @Success 200 {array} catalogDTO.EbookChapterResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/ebook-chapters [get]
func (h *Handler) ListEbookChapters(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	list, err := h.svc.ListEbookChapters(bookID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Obter capítulo ebook
// @Tags catalog
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Success 200 {object} catalogDTO.EbookChapterResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/ebook-chapters/{chapterId} [get]
func (h *Handler) GetEbookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	res, err := h.svc.GetEbookChapter(bookID, chapterID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Criar capítulo ebook
// @Tags catalog
// @Accept json
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param request body catalogDTO.CreateEbookChapterRequest true "Dados do capítulo"
// @Success 201 {object} catalogDTO.EbookChapterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/ebook-chapters [post]
func (h *Handler) CreateEbookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	var req catalogDTO.CreateEbookChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateEbookChapter(bookID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Atualizar capítulo ebook
// @Tags catalog
// @Accept json
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Param request body catalogDTO.UpdateEbookChapterRequest true "Campos a atualizar"
// @Success 200 {object} catalogDTO.EbookChapterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/ebook-chapters/{chapterId} [put]
func (h *Handler) UpdateEbookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	var req catalogDTO.UpdateEbookChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateEbookChapter(bookID, chapterID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir capítulo ebook
// @Tags catalog
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/ebook-chapters/{chapterId} [delete]
func (h *Handler) DeleteEbookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	if err := h.svc.DeleteEbookChapter(bookID, chapterID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Audiobook chapters ---

// @Summary Listar capítulos audiobook
// @Tags catalog
// @Produce json
// @Param bookId path string true "ID da obra"
// @Success 200 {array} catalogDTO.AudiobookChapterResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/audiobook-chapters [get]
func (h *Handler) ListAudiobookChapters(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	list, err := h.svc.ListAudiobookChapters(bookID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Obter capítulo audiobook
// @Tags catalog
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Success 200 {object} catalogDTO.AudiobookChapterResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/audiobook-chapters/{chapterId} [get]
func (h *Handler) GetAudiobookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	res, err := h.svc.GetAudiobookChapter(bookID, chapterID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Criar capítulo audiobook
// @Tags catalog
// @Accept json
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param request body catalogDTO.CreateAudiobookChapterRequest true "Dados do capítulo"
// @Success 201 {object} catalogDTO.AudiobookChapterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/audiobook-chapters [post]
func (h *Handler) CreateAudiobookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	var req catalogDTO.CreateAudiobookChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateAudiobookChapter(bookID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "obra não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Atualizar capítulo audiobook
// @Tags catalog
// @Accept json
// @Produce json
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Param request body catalogDTO.UpdateAudiobookChapterRequest true "Campos a atualizar"
// @Success 200 {object} catalogDTO.AudiobookChapterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/audiobook-chapters/{chapterId} [put]
func (h *Handler) UpdateAudiobookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	var req catalogDTO.UpdateAudiobookChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateAudiobookChapter(bookID, chapterID, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir capítulo audiobook
// @Tags catalog
// @Param bookId path string true "ID da obra"
// @Param chapterId path string true "ID do capítulo"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/books/{bookId}/audiobook-chapters/{chapterId} [delete]
func (h *Handler) DeleteAudiobookChapter(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	bookID := c.Param("bookId")
	chapterID := c.Param("chapterId")
	if err := h.svc.DeleteAudiobookChapter(bookID, chapterID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "capítulo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Coleções ---

// @Summary Criar coleção no catálogo
// @Description Agrupa obras do autor autenticado
// @Tags catalog
// @Accept json
// @Produce json
// @Param request body catalogDTO.CreateCollectionRequest true "Dados da coleção"
// @Success 201 {object} catalogDTO.CollectionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections [post]
func (h *Handler) CreateCollection(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	var req catalogDTO.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.CreateCollection(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// @Summary Listar coleções do autor
// @Tags catalog
// @Produce json
// @Success 200 {array} catalogDTO.CollectionResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections [get]
func (h *Handler) ListCollections(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	list, err := h.svc.ListCollections(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Obter coleção por ID
// @Tags catalog
// @Produce json
// @Param collectionId path string true "ID da coleção"
// @Success 200 {object} catalogDTO.CollectionResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections/{collectionId} [get]
func (h *Handler) GetCollection(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("collectionId")
	res, err := h.svc.GetCollection(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coleção não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Atualizar coleção
// @Tags catalog
// @Accept json
// @Produce json
// @Param collectionId path string true "ID da coleção"
// @Param request body catalogDTO.UpdateCollectionRequest true "Campos a atualizar"
// @Success 200 {object} catalogDTO.CollectionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections/{collectionId} [put]
func (h *Handler) UpdateCollection(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("collectionId")
	var req catalogDTO.UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.svc.UpdateCollection(id, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coleção não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Excluir coleção
// @Tags catalog
// @Param collectionId path string true "ID da coleção"
// @Success 204 "Sem corpo"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections/{collectionId} [delete]
func (h *Handler) DeleteCollection(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("collectionId")
	if err := h.svc.DeleteCollection(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coleção não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Substituir livros da coleção (ordem e rótulos de volume)
// @Tags catalog
// @Accept json
// @Produce json
// @Param collectionId path string true "ID da coleção"
// @Param request body catalogDTO.ReplaceCollectionBooksRequest true "IDs dos livros na ordem desejada"
// @Success 200 {object} catalogDTO.CollectionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/catalog/collections/{collectionId}/books [put]
func (h *Handler) ReplaceCollectionBooks(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return
	}
	id := c.Param("collectionId")
	var req catalogDTO.ReplaceCollectionBooksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.BookIDs == nil {
		req.BookIDs = []string{}
	}
	res, err := h.svc.ReplaceCollectionBooks(id, userID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coleção não encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
