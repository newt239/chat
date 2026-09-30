package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/customemoji"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type customEmojiRepository struct {
	client *ent.Client
}

func NewCustomEmojiRepository(client *ent.Client) domainrepository.CustomEmojiRepository {
	return &customEmojiRepository{client: client}
}

func (r *customEmojiRepository) FindByID(ctx context.Context, id string) (*entity.CustomEmoji, error) {
	emojiID, err := utils.ParseUUID(id, "custom emoji ID")
	if err != nil {
		return nil, err
	}
	found, err := transaction.ResolveClient(ctx, r.client).CustomEmoji.Get(ctx, emojiID)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return customEmojiToEntity(found), nil
}

func (r *customEmojiRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.CustomEmoji, error) {
	found, err := transaction.ResolveClient(ctx, r.client).CustomEmoji.Query().
		Where(customemoji.WorkspaceID(workspaceID)).
		Order(ent.Asc(customemoji.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.CustomEmoji, 0, len(found))
	for _, e := range found {
		result = append(result, customEmojiToEntity(e))
	}
	return result, nil
}

func (r *customEmojiRepository) Create(ctx context.Context, e *entity.CustomEmoji) error {
	creatorID, err := utils.ParseUUID(e.CreatedBy, "user ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).CustomEmoji.Create().
		SetWorkspaceID(e.WorkspaceID).
		SetName(e.Name).
		SetStorageKey(e.StorageKey).
		SetCreatedBy(creatorID)
	if e.ID != "" {
		id, err := utils.ParseUUID(e.ID, "custom emoji ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(id)
	}
	created, err := builder.Save(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrConflict
	}
	if err != nil {
		return err
	}
	e.ID = created.ID.String()
	e.CreatedAt = created.CreatedAt
	return nil
}

func (r *customEmojiRepository) Delete(ctx context.Context, id string) error {
	emojiID, err := utils.ParseUUID(id, "custom emoji ID")
	if err != nil {
		return err
	}
	err = transaction.ResolveClient(ctx, r.client).CustomEmoji.DeleteOneID(emojiID).Exec(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}

func customEmojiToEntity(e *ent.CustomEmoji) *entity.CustomEmoji {
	return &entity.CustomEmoji{
		ID:          e.ID.String(),
		WorkspaceID: e.WorkspaceID,
		Name:        e.Name,
		StorageKey:  e.StorageKey,
		CreatedBy:   e.CreatedBy.String(),
		CreatedAt:   e.CreatedAt,
	}
}
