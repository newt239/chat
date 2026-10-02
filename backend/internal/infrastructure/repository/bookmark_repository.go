package repository

import (
	"context"

	domerr "github.com/newt239/chat/internal/domain/errors"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/messagebookmark"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type bookmarkRepository struct {
	client *ent.Client
}

func NewBookmarkRepository(client *ent.Client) domainrepository.BookmarkRepository {
	return &bookmarkRepository{client: client}
}

func (r *bookmarkRepository) AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error {
	uid, err := utils.ParseUUID(bookmark.UserID, "user ID")
	if err != nil {
		return err
	}
	mid, err := utils.ParseUUID(bookmark.MessageID, "message ID")
	if err != nil {
		return err
	}
	saved, err := transaction.ResolveClient(ctx, r.client).MessageBookmark.Create().
		SetUserID(uid).
		SetMessageID(mid).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrBookmarkExists
	}
	if err != nil {
		return err
	}
	bookmark.CreatedAt = saved.CreatedAt
	return nil
}

func (r *bookmarkRepository) RemoveBookmark(ctx context.Context, userID, messageID string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.MessageBookmark.Delete().
		Where(
			messagebookmark.UserID(uid),
			messagebookmark.MessageID(mid),
		).
		Exec(ctx)

	return err
}

func (r *bookmarkRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.MessageBookmark, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	bookmarks, err := client.MessageBookmark.Query().
		Where(messagebookmark.UserID(uid)).
		WithMessage().
		Order(ent.Desc(messagebookmark.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.MessageBookmark, 0, len(bookmarks))
	for _, mb := range bookmarks {
		result = append(result, utils.MessageBookmarkToEntity(mb))
	}

	return result, nil
}
