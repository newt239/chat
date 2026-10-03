package draft

import (
	"context"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type SaveInput struct {
	Target domainrepository.DraftTarget
	Body   string
}

type Interactor struct {
	draftRepo        domainrepository.DraftRepository
	messageRepo      domainrepository.MessageRepository
	channelAccessSvc service.ChannelAccessService
}

func NewInteractor(draftRepo domainrepository.DraftRepository, messageRepo domainrepository.MessageRepository, channelAccessSvc service.ChannelAccessService) *Interactor {
	return &Interactor{draftRepo: draftRepo, messageRepo: messageRepo, channelAccessSvc: channelAccessSvc}
}

// Save は下書きを保存します。本文が空白だけなら削除して nil を返します
func (i *Interactor) Save(ctx context.Context, input SaveInput) (*entity.Draft, error) {
	if err := i.ensureTarget(ctx, input.Target); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Body) == "" {
		return nil, i.Delete(ctx, input.Target)
	}
	d := &entity.Draft{UserID: input.Target.UserID, ChannelID: input.Target.ChannelID, ParentID: input.Target.ParentID, Body: input.Body}
	if err := i.draftRepo.Upsert(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}
	return d, nil
}

// Get は下書きがなければ nil を返します。本人の下書きしか引けないため閲覧権限は確かめない
func (i *Interactor) Get(ctx context.Context, target domainrepository.DraftTarget) (*entity.Draft, error) {
	return i.draftRepo.Find(ctx, target)
}

func (i *Interactor) Delete(ctx context.Context, target domainrepository.DraftTarget) error {
	if err := i.draftRepo.Delete(ctx, target); err != nil {
		return fmt.Errorf("failed to delete draft: %w", err)
	}
	return nil
}

func (i *Interactor) List(ctx context.Context, userID, workspaceID string) ([]*entity.Draft, error) {
	return i.draftRepo.FindByWorkspace(ctx, userID, workspaceID)
}

// ensureTarget は投稿できるチャンネルで、返信先がそのチャンネルの削除されていないメッセージかを確かめます
func (i *Interactor) ensureTarget(ctx context.Context, target domainrepository.DraftTarget) error {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, target.ChannelID, target.UserID); err != nil {
		return err
	}
	if target.ParentID == nil {
		return nil
	}
	parent, err := i.messageRepo.FindByID(ctx, *target.ParentID)
	if err != nil {
		return fmt.Errorf("failed to load parent message: %w", err)
	}
	if !parent.CanBeRepliedIn(target.ChannelID) {
		return domerr.ErrParentMessageNotFound
	}
	return nil
}
