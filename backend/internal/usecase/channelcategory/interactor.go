package channelcategory

import (
	"context"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
)

var (
	ErrCategoryNotFound = domerr.New(domerr.ErrNotFound, "カテゴリが見つかりません")
	ErrInvalidOrder     = domerr.New(domerr.ErrValidation, "並び替えにはすべてのカテゴリを指定してください")
)

type Interactor struct {
	categoryRepo     domainrepository.ChannelCategoryRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	channelAccessSvc domainservice.ChannelAccessService
	txManager        domaintransaction.Manager
}

func New(
	categoryRepo domainrepository.ChannelCategoryRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	txManager domaintransaction.Manager,
) *Interactor {
	return &Interactor{categoryRepo: categoryRepo, workspaceRepo: workspaceRepo, channelAccessSvc: channelAccessSvc, txManager: txManager}
}

// findOwnCategory は他人のカテゴリを見つからないものとして扱います
func (i *Interactor) findOwnCategory(ctx context.Context, categoryID, userID string) (*entity.ChannelCategory, error) {
	category, err := i.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to load category: %w", err)
	}
	if category == nil || category.UserID != userID {
		return nil, ErrCategoryNotFound
	}
	return category, nil
}

func (i *Interactor) List(ctx context.Context, workspaceID, userID string) ([]*entity.ChannelCategory, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
	}
	categories, err := i.categoryRepo.FindByUser(ctx, userID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}
	return categories, nil
}

func (i *Interactor) Create(ctx context.Context, workspaceID, userID, name string) (*entity.ChannelCategory, error) {
	existing, err := i.List(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	category := &entity.ChannelCategory{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        strings.TrimSpace(name),
		Position:    len(existing),
		ChannelIDs:  []string{},
	}
	if err := i.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return category, nil
}

func (i *Interactor) Update(ctx context.Context, categoryID, userID, name string) (*entity.ChannelCategory, error) {
	category, err := i.findOwnCategory(ctx, categoryID, userID)
	if err != nil {
		return nil, err
	}
	category.Name = strings.TrimSpace(name)
	if err := i.categoryRepo.UpdateName(ctx, category.ID, category.Name); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return category, nil
}

func (i *Interactor) Delete(ctx context.Context, categoryID, userID string) error {
	if _, err := i.findOwnCategory(ctx, categoryID, userID); err != nil {
		return err
	}
	return i.categoryRepo.Delete(ctx, categoryID)
}

func (i *Interactor) Reorder(ctx context.Context, workspaceID, userID string, categoryIDs []string) error {
	return i.txManager.Do(ctx, func(txCtx context.Context) error {
		categories, err := i.categoryRepo.FindByUser(txCtx, userID, workspaceID)
		if err != nil {
			return fmt.Errorf("failed to load categories: %w", err)
		}
		if len(categories) != len(categoryIDs) {
			return ErrInvalidOrder
		}
		byID := make(map[string]bool, len(categories))
		for _, category := range categories {
			byID[category.ID] = true
		}
		for _, id := range categoryIDs {
			if !byID[id] {
				return ErrInvalidOrder
			}
		}
		return i.categoryRepo.UpdatePositions(txCtx, categoryIDs)
	})
}

// SetChannel はチャンネルをカテゴリに割り当てます。categoryID が nil なら割り当てを外します
func (i *Interactor) SetChannel(ctx context.Context, channelID, userID string, categoryID *string) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if categoryID != nil {
		category, err := i.findOwnCategory(ctx, *categoryID, userID)
		if err != nil {
			return err
		}
		if category.WorkspaceID != ch.WorkspaceID {
			return ErrCategoryNotFound
		}
	}
	return i.categoryRepo.SetChannel(ctx, userID, ch.ID, categoryID)
}
