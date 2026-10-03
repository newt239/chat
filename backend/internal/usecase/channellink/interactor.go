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
	ErrLinkNotFound = domerr.New(domerr.ErrNotFound, "関連リンクが見つかりません")
	ErrInvalidOrder = domerr.New(domerr.ErrValidation, "並び替えにはチャンネルのすべてのリンクを指定してください")
)

// LinkInput は作成ではチャンネル、編集ではリンクを ID で指します
type LinkInput struct {
	ID     string
	UserID string
	Title  string
	URL    string
}

type Interactor struct {
	linkRepo         domainrepository.ChannelLinkRepository
	channelAccessSvc domainservice.ChannelAccessService
	permissionSvc    domainservice.PermissionService
	txManager        domaintransaction.Manager
}

func New(
	linkRepo domainrepository.ChannelLinkRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	permissionSvc domainservice.PermissionService,
	txManager domaintransaction.Manager,
) *Interactor {
	return &Interactor{linkRepo: linkRepo, channelAccessSvc: channelAccessSvc, permissionSvc: permissionSvc, txManager: txManager}
}

// ensureEditable はチャンネルの閲覧権限と、権限設定の「関連リンクの編集」を確認します
func (i *Interactor) ensureEditable(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return err
	}
	_, err = i.permissionSvc.Ensure(ctx, ch.WorkspaceID, userID, entity.PermissionEditChannelLinks)
	return err
}

// findEditableLink はリンクを取得し、そのチャンネルの編集権限を確認します
func (i *Interactor) findEditableLink(ctx context.Context, linkID, userID string) (*entity.ChannelLink, error) {
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

// List はリンクと、閲覧者が編集できるかを返します
func (i *Interactor) List(ctx context.Context, channelID, userID string) ([]*entity.ChannelLink, bool, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return nil, false, err
	}
	links, err := i.linkRepo.FindByChannelID(ctx, ch.ID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load links: %w", err)
	}
	_, err = i.permissionSvc.Ensure(ctx, ch.WorkspaceID, userID, entity.PermissionEditChannelLinks)
	if errors.Is(err, domerr.ErrUnauthorized) {
		return links, false, nil
	}
	return links, err == nil, err
}

func (i *Interactor) Create(ctx context.Context, input LinkInput) (*entity.ChannelLink, error) {
	if err := i.ensureEditable(ctx, input.ID, input.UserID); err != nil {
		return nil, err
	}
	existing, err := i.linkRepo.FindByChannelID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load links: %w", err)
	}
	link := &entity.ChannelLink{
		ChannelID: input.ID,
		Title:     strings.TrimSpace(input.Title),
		URL:       strings.TrimSpace(input.URL),
		Position:  len(existing),
		CreatedBy: input.UserID,
	}
	if err := i.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}
	return link, nil
}

func (i *Interactor) Update(ctx context.Context, input LinkInput) (*entity.ChannelLink, error) {
	link, err := i.findEditableLink(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}
	link.Title = strings.TrimSpace(input.Title)
	link.URL = strings.TrimSpace(input.URL)
	if err := i.linkRepo.Update(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}
	return link, nil
}

func (i *Interactor) Delete(ctx context.Context, linkID, userID string) error {
	if _, err := i.findEditableLink(ctx, linkID, userID); err != nil {
		return err
	}
	if err := i.linkRepo.Delete(ctx, linkID); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}
	return nil
}

func (i *Interactor) Reorder(ctx context.Context, channelID, userID string, linkIDs []string) ([]*entity.ChannelLink, error) {
	if err := i.ensureEditable(ctx, channelID, userID); err != nil {
		return nil, err
	}
	var reordered []*entity.ChannelLink
	err := i.txManager.Do(ctx, func(txCtx context.Context) error {
		links, err := i.linkRepo.FindByChannelID(txCtx, channelID)
		if err != nil {
			return fmt.Errorf("failed to load links: %w", err)
		}
		if len(links) != len(linkIDs) {
			return ErrInvalidOrder
		}
		byID := make(map[string]*entity.ChannelLink, len(links))
		for _, link := range links {
			byID[link.ID] = link
		}
		for position, id := range linkIDs {
			link, ok := byID[id]
			if !ok {
				return ErrInvalidOrder
			}
			link.Position = position
			reordered = append(reordered, link)
		}
		return i.linkRepo.UpdatePositions(txCtx, linkIDs)
	})
	if err != nil {
		return nil, err
	}
	return reordered, nil
}
