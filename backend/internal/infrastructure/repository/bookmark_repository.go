package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagebookmark"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type bookmarkRepository struct {
	client *ent.Client
}

func NewBookmarkRepository(client *ent.Client) domainrepository.BookmarkRepository {
	return &bookmarkRepository{client: client}
}

func (r *bookmarkRepository) AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error {
	uid, err := parseUUID(bookmark.UserID, "user ID")
	if err != nil {
		return err
	}
	mid, err := parseUUID(bookmark.MessageID, "message ID")
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
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	mid, err := parseUUID(messageID, "message ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).MessageBookmark.Delete().
		Where(messagebookmark.UserID(uid), messagebookmark.MessageID(mid)).
		Exec(ctx)
	return err
}

func (r *bookmarkRepository) FindByUserID(ctx context.Context, userID, workspaceID string) ([]*entity.MessageBookmark, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	bookmarks, err := transaction.ResolveClient(ctx, r.client).MessageBookmark.Query().
		Where(messagebookmark.UserID(uid), messagebookmark.HasMessageWith(message.DeletedAtIsNil(), message.HasChannelWith(channel.WorkspaceID(workspaceID)))).
		WithMessage().
		Order(ent.Desc(messagebookmark.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(bookmarks, func(mb *ent.MessageBookmark) *entity.MessageBookmark {
		return &entity.MessageBookmark{
			UserID:    mb.UserID.String(),
			MessageID: mb.MessageID.String(),
			Message:   messageToEntity(mb.Edges.Message),
			CreatedAt: mb.CreatedAt,
		}
	}), nil
}
