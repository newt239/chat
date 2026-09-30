package channelcategory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
)

var (
	ErrUnauthorized     = errors.New("このワークスペースのメンバーではありません")
	ErrCategoryNotFound = errors.New("カテゴリが見つかりません")
	ErrInvalidOrder     = fmt.Errorf("%w: 並び替えにはすべてのカテゴリを指定してください", domerr.ErrValidation)
)

type UseCase interface {
	List(ctx context.Context, input ListInput) ([]CategoryOutput, error)
	Create(ctx context.Context, input CreateInput) (*CategoryOutput, error)
	Update(ctx context.Context, input UpdateInput) (*CategoryOutput, error)
	Delete(ctx context.Context, input DeleteInput) error
	Reorder(ctx context.Context, input ReorderInput) ([]CategoryOutput, error)
	SetChannel(ctx context.Context, input SetChannelInput) error
}

type interactor struct {
	categoryRepo     domainrepository.ChannelCategoryRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	channelAccessSvc domainservice.ChannelAccessService
	txManager        domaintransaction.Manager
}

func NewInteractor(
	categoryRepo domainrepository.ChannelCategoryRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	txManager domaintransaction.Manager,
) UseCase {
	return &interactor{
		categoryRepo:     categoryRepo,
		workspaceRepo:    workspaceRepo,
		channelAccessSvc: channelAccessSvc,
		txManager:        txManager,
	}
}

func (i *interactor) ensureWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return ErrUnauthorized
	}
	return nil
}

// findOwnCategory は他人のカテゴリを見つからないものとして扱います
func (i *interactor) findOwnCategory(ctx context.Context, categoryID, userID string) (*entity.ChannelCategory, error) {
	category, err := i.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to load category: %w", err)
	}
	if category == nil || category.UserID != userID {
		return nil, ErrCategoryNotFound
	}
	return category, nil
}

func (i *interactor) List(ctx context.Context, input ListInput) ([]CategoryOutput, error) {
	if err := i.ensureWorkspaceMember(ctx, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}
	categories, err := i.categoryRepo.FindByUser(ctx, input.UserID, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}
	return toOutputs(categories), nil
}

func (i *interactor) Create(ctx context.Context, input CreateInput) (*CategoryOutput, error) {
	if err := i.ensureWorkspaceMember(ctx, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}
	existing, err := i.categoryRepo.FindByUser(ctx, input.UserID, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}
	category := &entity.ChannelCategory{
		UserID:      input.UserID,
		WorkspaceID: input.WorkspaceID,
		Name:        strings.TrimSpace(input.Name),
		Position:    len(existing),
		ChannelIDs:  []string{},
	}
	if err := i.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	out := toOutput(category)
	return &out, nil
}

func (i *interactor) Update(ctx context.Context, input UpdateInput) (*CategoryOutput, error) {
	category, err := i.findOwnCategory(ctx, input.CategoryID, input.UserID)
	if err != nil {
		return nil, err
	}
	category.Name = strings.TrimSpace(input.Name)
	if err := i.categoryRepo.UpdateName(ctx, category.ID, category.Name); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	out := toOutput(category)
	return &out, nil
}

func (i *interactor) Delete(ctx context.Context, input DeleteInput) error {
	if _, err := i.findOwnCategory(ctx, input.CategoryID, input.UserID); err != nil {
		return err
	}
	return i.txManager.Do(ctx, func(txCtx context.Context) error {
		return i.categoryRepo.Delete(txCtx, input.CategoryID)
	})
}

func (i *interactor) Reorder(ctx context.Context, input ReorderInput) ([]CategoryOutput, error) {
	var reordered []*entity.ChannelCategory
	err := i.txManager.Do(ctx, func(txCtx context.Context) error {
		categories, err := i.categoryRepo.FindByUser(txCtx, input.UserID, input.WorkspaceID)
		if err != nil {
			return fmt.Errorf("failed to load categories: %w", err)
		}
		if len(categories) != len(input.CategoryIDs) {
			return ErrInvalidOrder
		}
		byID := make(map[string]*entity.ChannelCategory, len(categories))
		for _, category := range categories {
			byID[category.ID] = category
		}
		for position, id := range input.CategoryIDs {
			category, ok := byID[id]
			if !ok {
				return ErrInvalidOrder
			}
			category.Position = position
			reordered = append(reordered, category)
		}
		return i.categoryRepo.UpdatePositions(txCtx, input.CategoryIDs)
	})
	if err != nil {
		return nil, err
	}
	return toOutputs(reordered), nil
}

func (i *interactor) SetChannel(ctx context.Context, input SetChannelInput) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return err
	}
	if input.CategoryID != nil {
		category, err := i.findOwnCategory(ctx, *input.CategoryID, input.UserID)
		if err != nil {
			return err
		}
		if category.WorkspaceID != ch.WorkspaceID {
			return ErrCategoryNotFound
		}
	}
	return i.txManager.Do(ctx, func(txCtx context.Context) error {
		return i.categoryRepo.SetChannel(txCtx, input.UserID, ch.ID, input.CategoryID)
	})
}

func toOutput(category *entity.ChannelCategory) CategoryOutput {
	return CategoryOutput{
		ID:         category.ID,
		Name:       category.Name,
		Position:   category.Position,
		ChannelIDs: category.ChannelIDs,
	}
}

func toOutputs(categories []*entity.ChannelCategory) []CategoryOutput {
	outputs := make([]CategoryOutput, 0, len(categories))
	for _, category := range categories {
		outputs = append(outputs, toOutput(category))
	}
	return outputs
}
