package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/customemoji"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type customEmojiRepository struct {
	client *ent.Client
}

func NewCustomEmojiRepository(client *ent.Client) domainrepository.CustomEmojiRepository {
	return &customEmojiRepository{client: client}
}

func (r *customEmojiRepository) FindByID(ctx context.Context, id string) (*entity.CustomEmoji, error) {
	emojiID, err := parseUUID(id, "custom emoji ID")
	if err != nil {
		return nil, err
	}
	found, err := orNil(transaction.ResolveClient(ctx, r.client).CustomEmoji.Get(ctx, emojiID))
	if found == nil {
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
	return convertAll(found, customEmojiToEntity), nil
}

func (r *customEmojiRepository) Create(ctx context.Context, e *entity.CustomEmoji) error {
	creatorID, err := parseUUID(e.CreatedBy, "user ID")
	if err != nil {
		return err
	}
	id, err := parseUUID(e.ID, "custom emoji ID")
	if err != nil {
		return err
	}
	err = transaction.ResolveClient(ctx, r.client).CustomEmoji.Create().
		SetID(id).
		SetWorkspaceID(e.WorkspaceID).
		SetName(e.Name).
		SetStorageKey(e.StorageKey).
		SetCreatedBy(creatorID).
		Exec(ctx)
	if ent.IsConstraintError(err) {
		return domerr.ErrCustomEmojiNameExists
	}
	return err
}

func (r *customEmojiRepository) Delete(ctx context.Context, id string) error {
	emojiID, err := parseUUID(id, "custom emoji ID")
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
	}
}
