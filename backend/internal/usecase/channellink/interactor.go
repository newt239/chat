package channellink

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
	ErrUnauthorized = errors.New("関連リンクを編集する権限がありません")
	ErrLinkNotFound = errors.New("関連リンクが見つかりません")
	ErrInvalidOrder = fmt.Errorf("%w: 並び替えにはチャンネルのすべてのリンクを指定してください", domerr.ErrValidation)
)

type UseCase interface {
	List(ctx context.Context, input ListInput) (*ListOutput, error)
	Create(ctx context.Context, input CreateInput) (*LinkOutput, error)
	Update(ctx context.Context, input UpdateInput) (*LinkOutput, error)
	Delete(ctx context.Context, input DeleteInput) error
	Reorder(ctx context.Context, input ReorderInput) ([]LinkOutput, error)
}

type interactor struct {
	linkRepo         domainrepository.ChannelLinkRepository
	channelAccessSvc domainservice.ChannelAccessService
	permissionSvc    domainservice.PermissionService
	txManager        domaintransaction.Manager
}

func NewInteractor(
	linkRepo domainrepository.ChannelLinkRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	permissionSvc domainservice.PermissionService,
	txManager domaintransaction.Manager,
) UseCase {
	return &interactor{
		linkRepo:         linkRepo,
		channelAccessSvc: channelAccessSvc,
		permissionSvc:    permissionSvc,
		txManager:        txManager,
	}
}

// canEdit は権限設定の「関連リンクの編集」が許可されているかを返します
func (i *interactor) canEdit(ctx context.Context, ch *entity.Channel, userID string) (bool, error) {
	_, err := i.permissionSvc.Ensure(ctx, ch.WorkspaceID, userID, entity.PermissionEditChannelLinks)
	if errors.Is(err, domerr.ErrUnauthorized) {
		return false, nil
	}
	return err == nil, err
}

// ensureEditable はチャンネルの閲覧権限と関連リンクの編集権限を確認します
func (i *interactor) ensureEditable(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return err
	}
	editable, err := i.canEdit(ctx, ch, userID)
	if err != nil {
		return err
	}
	if !editable {
		return ErrUnauthorized
	}
	return nil
}

// findEditableLink はリンクを取得し、そのチャンネルの編集権限を確認します
func (i *interactor) findEditableLink(ctx context.Context, linkID, userID string) (*entity.ChannelLink, error) {
	link, err := i.linkRepo.FindByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to load link: %w", err)
	}
	if link == nil {
		return nil, ErrLinkNotFound
	}
	if err := i.ensureEditable(ctx, link.ChannelID, userID); err != nil {
		return nil, err
	}
	return link, nil
}

func (i *interactor) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	links, err := i.linkRepo.FindByChannelID(ctx, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load links: %w", err)
	}
	editable, err := i.canEdit(ctx, ch, input.UserID)
	if err != nil {
		return nil, err
	}
	return &ListOutput{Links: toOutputs(links), CanEdit: editable}, nil
}

func (i *interactor) Create(ctx context.Context, input CreateInput) (*LinkOutput, error) {
	if err := i.ensureEditable(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}
	existing, err := i.linkRepo.FindByChannelID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load links: %w", err)
	}

	link := &entity.ChannelLink{
		ChannelID: input.ChannelID,
		Title:     strings.TrimSpace(input.Title),
		URL:       strings.TrimSpace(input.URL),
		Position:  len(existing),
		CreatedBy: input.UserID,
	}
	if err := i.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}
	out := toOutput(link)
	return &out, nil
}

func (i *interactor) Update(ctx context.Context, input UpdateInput) (*LinkOutput, error) {
	link, err := i.findEditableLink(ctx, input.LinkID, input.UserID)
	if err != nil {
		return nil, err
	}
	link.Title = strings.TrimSpace(input.Title)
	link.URL = strings.TrimSpace(input.URL)
	if err := i.linkRepo.Update(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}
	out := toOutput(link)
	return &out, nil
}

func (i *interactor) Delete(ctx context.Context, input DeleteInput) error {
	if _, err := i.findEditableLink(ctx, input.LinkID, input.UserID); err != nil {
		return err
	}
	if err := i.linkRepo.Delete(ctx, input.LinkID); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}
	return nil
}

func (i *interactor) Reorder(ctx context.Context, input ReorderInput) ([]LinkOutput, error) {
	if err := i.ensureEditable(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}

	var reordered []*entity.ChannelLink
	err := i.txManager.Do(ctx, func(txCtx context.Context) error {
		links, err := i.linkRepo.FindByChannelID(txCtx, input.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to load links: %w", err)
		}
		if len(links) != len(input.LinkIDs) {
			return ErrInvalidOrder
		}
		byID := make(map[string]*entity.ChannelLink, len(links))
		for _, link := range links {
			byID[link.ID] = link
		}
		for position, id := range input.LinkIDs {
			link, ok := byID[id]
			if !ok {
				return ErrInvalidOrder
			}
			link.Position = position
			reordered = append(reordered, link)
		}
		return i.linkRepo.UpdatePositions(txCtx, input.LinkIDs)
	})
	if err != nil {
		return nil, err
	}
	return toOutputs(reordered), nil
}

func toOutput(link *entity.ChannelLink) LinkOutput {
	return LinkOutput{
		ID:        link.ID,
		ChannelID: link.ChannelID,
		Title:     link.Title,
		URL:       link.URL,
		Position:  link.Position,
		CreatedBy: link.CreatedBy,
		CreatedAt: link.CreatedAt,
		UpdatedAt: link.UpdatedAt,
	}
}

func toOutputs(links []*entity.ChannelLink) []LinkOutput {
	outputs := make([]LinkOutput, 0, len(links))
	for _, link := range links {
		outputs = append(outputs, toOutput(link))
	}
	return outputs
}
