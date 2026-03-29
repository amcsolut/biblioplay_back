package feed

import (
	"errors"
	"time"

	feedDTO "api-backend-infinitrum/api/v1/dto/feed"
	feedmodel "api-backend-infinitrum/internal/models/feed"
	"api-backend-infinitrum/internal/models/profile"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) communityExists(id string) error {
	var n int64
	if err := s.db.Model(&profile.Community{}).Where("id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func validReactionType(t string) bool {
	switch t {
	case feedmodel.ReactionLike, feedmodel.ReactionLove, feedmodel.ReactionLaugh,
		feedmodel.ReactionSad, feedmodel.ReactionAngry, feedmodel.ReactionWow:
		return true
	default:
		return false
	}
}

func validTargetType(t string) bool {
	switch t {
	case feedmodel.TargetTypePost, feedmodel.TargetTypeComment, feedmodel.TargetTypeReply:
		return true
	default:
		return false
	}
}

// --- Posts ---

func (s *Service) ListPosts(communityID string, page, limit int) ([]feedDTO.PostResponse, error) {
	if err := s.communityExists(communityID); err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var list []feedmodel.CommunityPost
	err := s.db.Where("community_id = ? AND deleted_at IS NULL", communityID).
		Order("pinned DESC, created_at DESC").
		Limit(limit).Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]feedDTO.PostResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPostResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) CreatePost(communityID, authorUserID string, req *feedDTO.CreatePostRequest) (*feedDTO.PostResponse, error) {
	if err := s.communityExists(communityID); err != nil {
		return nil, err
	}
	p := &feedmodel.CommunityPost{
		CommunityID:  communityID,
		AuthorUserID: authorUserID,
		Body:         req.Body,
	}
	if req.Status != "" {
		p.Status = req.Status
	}
	if req.Pinned != nil {
		p.Pinned = *req.Pinned
	}
	if err := s.db.Create(p).Error; err != nil {
		return nil, err
	}
	return toPostResponse(p), nil
}

func (s *Service) GetPost(communityID, postID string) (*feedDTO.PostResponse, error) {
	var p feedmodel.CommunityPost
	err := s.db.Where("id = ? AND community_id = ? AND deleted_at IS NULL", postID, communityID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return toPostResponse(&p), nil
}

func (s *Service) UpdatePost(communityID, postID, userID string, req *feedDTO.UpdatePostRequest) (*feedDTO.PostResponse, error) {
	var p feedmodel.CommunityPost
	if err := s.db.Where("id = ? AND community_id = ? AND deleted_at IS NULL", postID, communityID).First(&p).Error; err != nil {
		return nil, err
	}
	if p.AuthorUserID != userID {
		return nil, errors.New("apenas o autor pode editar o post")
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Pinned != nil {
		updates["pinned"] = *req.Pinned
	}
	if err := s.db.Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&p, "id = ?", p.ID).Error; err != nil {
		return nil, err
	}
	return toPostResponse(&p), nil
}

func (s *Service) DeletePost(communityID, postID, userID string) error {
	var p feedmodel.CommunityPost
	if err := s.db.Where("id = ? AND community_id = ? AND deleted_at IS NULL", postID, communityID).First(&p).Error; err != nil {
		return err
	}
	if p.AuthorUserID != userID {
		return errors.New("apenas o autor pode excluir o post")
	}
	now := time.Now()
	return s.db.Model(&p).Updates(map[string]interface{}{
		"deleted_at": now,
		"status":     feedmodel.PostStatusDeleted,
		"updated_at": now,
	}).Error
}

func toPostResponse(p *feedmodel.CommunityPost) *feedDTO.PostResponse {
	return &feedDTO.PostResponse{
		ID:           p.ID,
		CommunityID:  p.CommunityID,
		AuthorUserID: p.AuthorUserID,
		Body:         p.Body,
		Status:       p.Status,
		Pinned:       p.Pinned,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}
}

func (s *Service) postInCommunity(communityID, postID string) (*feedmodel.CommunityPost, error) {
	var p feedmodel.CommunityPost
	err := s.db.Where("id = ? AND community_id = ? AND deleted_at IS NULL", postID, communityID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// --- Media ---

func (s *Service) ListPostMedia(communityID, postID string) ([]feedDTO.PostMediaResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var list []feedmodel.PostMedia
	if err := s.db.Where("post_id = ?", postID).Order("position ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]feedDTO.PostMediaResponse, 0, len(list))
	for i := range list {
		out = append(out, *toMediaResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) CreatePostMedia(communityID, postID, userID string, req *feedDTO.CreatePostMediaRequest) (*feedDTO.PostMediaResponse, error) {
	p, err := s.postInCommunity(communityID, postID)
	if err != nil {
		return nil, err
	}
	if p.AuthorUserID != userID {
		return nil, errors.New("apenas o autor pode adicionar mídia")
	}
	if req.MediaType != feedmodel.MediaTypeImage && req.MediaType != feedmodel.MediaTypeVideo {
		return nil, errors.New("media_type inválido (image ou video)")
	}
	m := &feedmodel.PostMedia{
		PostID:          postID,
		MediaType:       req.MediaType,
		StorageKey:      req.StorageKey,
		MediaURL:        req.MediaURL,
		ThumbnailURL:    req.ThumbnailURL,
		ThumbnailKey:    req.ThumbnailKey,
		Position:        req.Position,
		Width:           req.Width,
		Height:          req.Height,
		DurationSeconds: req.DurationSeconds,
		MimeType:        req.MimeType,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return toMediaResponse(m), nil
}

func (s *Service) GetPostMedia(communityID, postID, mediaID string) (*feedDTO.PostMediaResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var m feedmodel.PostMedia
	if err := s.db.Where("id = ? AND post_id = ?", mediaID, postID).First(&m).Error; err != nil {
		return nil, err
	}
	return toMediaResponse(&m), nil
}

func (s *Service) UpdatePostMedia(communityID, postID, mediaID, userID string, req *feedDTO.UpdatePostMediaRequest) (*feedDTO.PostMediaResponse, error) {
	p, err := s.postInCommunity(communityID, postID)
	if err != nil {
		return nil, err
	}
	if p.AuthorUserID != userID {
		return nil, errors.New("apenas o autor pode editar mídia")
	}
	var m feedmodel.PostMedia
	if err := s.db.Where("id = ? AND post_id = ?", mediaID, postID).First(&m).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.MediaType != nil {
		if *req.MediaType != feedmodel.MediaTypeImage && *req.MediaType != feedmodel.MediaTypeVideo {
			return nil, errors.New("media_type inválido")
		}
		updates["media_type"] = *req.MediaType
	}
	if req.StorageKey != nil {
		updates["storage_key"] = *req.StorageKey
	}
	if req.MediaURL != nil {
		updates["media_url"] = *req.MediaURL
	}
	if req.ThumbnailURL != nil {
		updates["thumbnail_url"] = *req.ThumbnailURL
	}
	if req.ThumbnailKey != nil {
		updates["thumbnail_key"] = *req.ThumbnailKey
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.Width != nil {
		updates["width"] = *req.Width
	}
	if req.Height != nil {
		updates["height"] = *req.Height
	}
	if req.DurationSeconds != nil {
		updates["duration_seconds"] = *req.DurationSeconds
	}
	if req.MimeType != nil {
		updates["mime_type"] = *req.MimeType
	}
	if len(updates) == 0 {
		return toMediaResponse(&m), nil
	}
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&m).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&m, "id = ?", m.ID).Error; err != nil {
		return nil, err
	}
	return toMediaResponse(&m), nil
}

func (s *Service) DeletePostMedia(communityID, postID, mediaID, userID string) error {
	p, err := s.postInCommunity(communityID, postID)
	if err != nil {
		return err
	}
	if p.AuthorUserID != userID {
		return errors.New("apenas o autor pode excluir mídia")
	}
	res := s.db.Where("id = ? AND post_id = ?", mediaID, postID).Delete(&feedmodel.PostMedia{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toMediaResponse(m *feedmodel.PostMedia) *feedDTO.PostMediaResponse {
	return &feedDTO.PostMediaResponse{
		ID:              m.ID,
		PostID:          m.PostID,
		MediaType:       m.MediaType,
		StorageKey:      m.StorageKey,
		MediaURL:        m.MediaURL,
		ThumbnailURL:    m.ThumbnailURL,
		ThumbnailKey:    m.ThumbnailKey,
		Position:        m.Position,
		Width:           m.Width,
		Height:          m.Height,
		DurationSeconds: m.DurationSeconds,
		MimeType:        m.MimeType,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// --- Comments ---

func (s *Service) ListComments(communityID, postID string) ([]feedDTO.CommentResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var list []feedmodel.PostComment
	err := s.db.Where("post_id = ? AND deleted_at IS NULL", postID).Order("created_at ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]feedDTO.CommentResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCommentResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) CreateComment(communityID, postID, authorUserID string, req *feedDTO.CreateCommentRequest) (*feedDTO.CommentResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	c := &feedmodel.PostComment{
		PostID:       postID,
		AuthorUserID: authorUserID,
		Body:         req.Body,
	}
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return toCommentResponse(c), nil
}

func (s *Service) GetComment(communityID, postID, commentID string) (*feedDTO.CommentResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var c feedmodel.PostComment
	if err := s.db.Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).First(&c).Error; err != nil {
		return nil, err
	}
	return toCommentResponse(&c), nil
}

func (s *Service) UpdateComment(communityID, postID, commentID, userID string, req *feedDTO.UpdateCommentRequest) (*feedDTO.CommentResponse, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var c feedmodel.PostComment
	if err := s.db.Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).First(&c).Error; err != nil {
		return nil, err
	}
	if c.AuthorUserID != userID {
		return nil, errors.New("apenas o autor pode editar o comentário")
	}
	if err := s.db.Model(&c).Updates(map[string]interface{}{
		"body":       req.Body,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&c, "id = ?", c.ID).Error; err != nil {
		return nil, err
	}
	return toCommentResponse(&c), nil
}

func (s *Service) DeleteComment(communityID, postID, commentID, userID string) error {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return err
	}
	var c feedmodel.PostComment
	if err := s.db.Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).First(&c).Error; err != nil {
		return err
	}
	if c.AuthorUserID != userID {
		return errors.New("apenas o autor pode excluir o comentário")
	}
	now := time.Now()
	return s.db.Model(&c).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error
}

func toCommentResponse(c *feedmodel.PostComment) *feedDTO.CommentResponse {
	return &feedDTO.CommentResponse{
		ID:           c.ID,
		PostID:       c.PostID,
		AuthorUserID: c.AuthorUserID,
		Body:         c.Body,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		DeletedAt:    c.DeletedAt,
	}
}

// --- Replies ---

func (s *Service) ListReplies(communityID, postID, commentID string) ([]feedDTO.ReplyResponse, error) {
	if _, err := s.getCommentInPost(communityID, postID, commentID); err != nil {
		return nil, err
	}
	var list []feedmodel.CommentReply
	err := s.db.Where("comment_id = ? AND deleted_at IS NULL", commentID).Order("created_at ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]feedDTO.ReplyResponse, 0, len(list))
	for i := range list {
		out = append(out, *toReplyResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) getCommentInPost(communityID, postID, commentID string) (*feedmodel.PostComment, error) {
	if _, err := s.postInCommunity(communityID, postID); err != nil {
		return nil, err
	}
	var c feedmodel.PostComment
	if err := s.db.Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) CreateReply(communityID, postID, commentID, authorUserID string, req *feedDTO.CreateReplyRequest) (*feedDTO.ReplyResponse, error) {
	if _, err := s.getCommentInPost(communityID, postID, commentID); err != nil {
		return nil, err
	}
	if req.ParentReplyID != nil && *req.ParentReplyID != "" {
		var parent feedmodel.CommentReply
		if err := s.db.Where("id = ? AND comment_id = ? AND deleted_at IS NULL", *req.ParentReplyID, commentID).First(&parent).Error; err != nil {
			return nil, errors.New("parent_reply_id inválido")
		}
	}
	r := &feedmodel.CommentReply{
		CommentID:      commentID,
		ParentReplyID:  req.ParentReplyID,
		AuthorUserID:   authorUserID,
		Body:           req.Body,
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return toReplyResponse(r), nil
}

func (s *Service) GetReply(communityID, postID, commentID, replyID string) (*feedDTO.ReplyResponse, error) {
	if _, err := s.getCommentInPost(communityID, postID, commentID); err != nil {
		return nil, err
	}
	var r feedmodel.CommentReply
	if err := s.db.Where("id = ? AND comment_id = ? AND deleted_at IS NULL", replyID, commentID).First(&r).Error; err != nil {
		return nil, err
	}
	return toReplyResponse(&r), nil
}

func (s *Service) UpdateReply(communityID, postID, commentID, replyID, userID string, req *feedDTO.UpdateReplyRequest) (*feedDTO.ReplyResponse, error) {
	if _, err := s.getCommentInPost(communityID, postID, commentID); err != nil {
		return nil, err
	}
	var r feedmodel.CommentReply
	if err := s.db.Where("id = ? AND comment_id = ? AND deleted_at IS NULL", replyID, commentID).First(&r).Error; err != nil {
		return nil, err
	}
	if r.AuthorUserID != userID {
		return nil, errors.New("apenas o autor pode editar a resposta")
	}
	if err := s.db.Model(&r).Updates(map[string]interface{}{
		"body":       req.Body,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&r, "id = ?", r.ID).Error; err != nil {
		return nil, err
	}
	return toReplyResponse(&r), nil
}

func (s *Service) DeleteReply(communityID, postID, commentID, replyID, userID string) error {
	if _, err := s.getCommentInPost(communityID, postID, commentID); err != nil {
		return err
	}
	var r feedmodel.CommentReply
	if err := s.db.Where("id = ? AND comment_id = ? AND deleted_at IS NULL", replyID, commentID).First(&r).Error; err != nil {
		return err
	}
	if r.AuthorUserID != userID {
		return errors.New("apenas o autor pode excluir a resposta")
	}
	now := time.Now()
	return s.db.Model(&r).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error
}

func toReplyResponse(r *feedmodel.CommentReply) *feedDTO.ReplyResponse {
	return &feedDTO.ReplyResponse{
		ID:            r.ID,
		CommentID:     r.CommentID,
		ParentReplyID: r.ParentReplyID,
		AuthorUserID:  r.AuthorUserID,
		Body:          r.Body,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		DeletedAt:     r.DeletedAt,
	}
}

// --- Reactions ---

func (s *Service) ListReactions(targetType, targetID string) ([]feedDTO.ReactionResponse, error) {
	if !validTargetType(targetType) {
		return nil, errors.New("target_type inválido")
	}
	var list []feedmodel.Reaction
	if err := s.db.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]feedDTO.ReactionResponse, 0, len(list))
	for i := range list {
		out = append(out, *toReactionResponse(&list[i]))
	}
	return out, nil
}

func (s *Service) UpsertReaction(userID string, req *feedDTO.UpsertReactionRequest) (*feedDTO.ReactionResponse, error) {
	if !validTargetType(req.TargetType) {
		return nil, errors.New("target_type inválido")
	}
	if !validReactionType(req.Type) {
		return nil, errors.New("type de reação inválido")
	}

	var existing feedmodel.Reaction
	err := s.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, req.TargetType, req.TargetID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r := &feedmodel.Reaction{
			UserID:     userID,
			TargetType: req.TargetType,
			TargetID:   req.TargetID,
			Type:       req.Type,
		}
		if err := s.db.Create(r).Error; err != nil {
			return nil, err
		}
		return toReactionResponse(r), nil
	}
	existing.Type = req.Type
	existing.UpdatedAt = time.Now()
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return toReactionResponse(&existing), nil
}

func (s *Service) DeleteReaction(reactionID, userID string) error {
	var r feedmodel.Reaction
	if err := s.db.Where("id = ?", reactionID).First(&r).Error; err != nil {
		return err
	}
	if r.UserID != userID {
		return errors.New("apenas o autor da reação pode removê-la")
	}
	return s.db.Delete(&r).Error
}

func toReactionResponse(r *feedmodel.Reaction) *feedDTO.ReactionResponse {
	return &feedDTO.ReactionResponse{
		ID:         r.ID,
		UserID:     r.UserID,
		TargetType: r.TargetType,
		TargetID:   r.TargetID,
		Type:       r.Type,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
