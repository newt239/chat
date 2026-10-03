package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelcategory"
	"github.com/newt239/chat/ent/channelcategoryitem"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelCategoryRepository struct {
	client *ent.Client
}

func NewChannelCategoryRepository(client *ent.Client) domainrepository.ChannelCategoryRepository {
	return &channelCategoryRepository{client: client}
}

func withCategoryEdges(q *ent.ChannelCategoryQuery) *ent.ChannelCategoryQuery {
	return q.WithItems()
}

func (r *channelCategoryRepository) FindByID(ctx context.Context, id string) (*entity.ChannelCategory, error) {
	categoryID, err := utils.ParseUUID(id, "category ID")
	if err != nil {
		return nil, err
	}
	client := transaction.ResolveClient(ctx, r.client)
	category, err := withCategoryEdges(client.ChannelCategory.Query().Where(channelcategory.ID(categoryID))).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return channelCategoryToEntity(category), nil
}

func (r *channelCategoryRepository) FindByUser(ctx context.Context, userID string, workspaceID string) ([]*entity.ChannelCategory, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	client := transaction.ResolveClient(ctx, r.client)
	categories, err := withCategoryEdges(client.ChannelCategory.Query().
		Where(
			channelcategory.UserID(uid),
			channelcategory.WorkspaceID(workspaceID),
		)).
		Order(ent.Asc(channelcategory.FieldPosition), ent.Asc(channelcategory.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.ChannelCategory, 0, len(categories))
	for _, category := range categories {
		result = append(result, channelCategoryToEntity(category))
	}
	return result, nil
}

func (r *channelCategoryRepository) Create(ctx context.Context, category *entity.ChannelCategory) error {
	uid, err := utils.ParseUUID(category.UserID, "user ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	created, err := client.ChannelCategory.Create().
		SetUserID(uid).
		SetWorkspaceID(category.WorkspaceID).
		SetName(category.Name).
		SetPosition(category.Position).
		Save(ctx)
	if err != nil {
		return err
	}
	category.ID = created.ID.String()
	return nil
}

func (r *channelCategoryRepository) UpdateName(ctx context.Context, id string, name string) error {
	categoryID, err := utils.ParseUUID(id, "category ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ChannelCategory.UpdateOneID(categoryID).SetName(name).Exec(ctx)
}

func (r *channelCategoryRepository) Delete(ctx context.Context, id string) error {
	categoryID, err := utils.ParseUUID(id, "category ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	if _, err := client.ChannelCategoryItem.Delete().
		Where(channelcategoryitem.CategoryID(categoryID)).
		Exec(ctx); err != nil {
		return err
	}
	return client.ChannelCategory.DeleteOneID(categoryID).Exec(ctx)
}

func (r *channelCategoryRepository) UpdatePositions(ctx context.Context, categoryIDs []string) error {
	client := transaction.ResolveClient(ctx, r.client)
	for position, id := range categoryIDs {
		categoryID, err := utils.ParseUUID(id, "category ID")
		if err != nil {
			return err
		}
		if err := client.ChannelCategory.UpdateOneID(categoryID).SetPosition(position).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *channelCategoryRepository) SetChannel(ctx context.Context, userID string, channelID string, categoryID *string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	if _, err := client.ChannelCategoryItem.Delete().
		Where(
			channelcategoryitem.UserID(uid),
			channelcategoryitem.ChannelID(cid),
		).
		Exec(ctx); err != nil {
		return err
	}
	if categoryID == nil {
		return nil
	}
	catID, err := utils.ParseUUID(*categoryID, "category ID")
	if err != nil {
		return err
	}
	return client.ChannelCategoryItem.Create().
		SetCategoryID(catID).
		SetUserID(uid).
		SetChannelID(cid).
		Exec(ctx)
}

func channelCategoryToEntity(category *ent.ChannelCategory) *entity.ChannelCategory {
	result := &entity.ChannelCategory{
		ID:          category.ID.String(),
		Name:        category.Name,
		Position:    category.Position,
		UserID:      category.UserID.String(),
		WorkspaceID: category.WorkspaceID,
		ChannelIDs:  make([]string, 0, len(category.Edges.Items)),
	}
	for _, item := range category.Edges.Items {
		result.ChannelIDs = append(result.ChannelIDs, item.ChannelID.String())
	}
	return result
}
