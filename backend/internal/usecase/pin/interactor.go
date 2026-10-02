package pin

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

type PinUseCase interface {
	PinMessage(ctx context.Context, input PinMessageInput) error
	UnpinMessage(ctx context.Context, input UnpinMessageInput) error
	ListPins(ctx context.Context, input ListPinsInput) (*ListPinsOutput, error)
}

type interactor struct {
	pinRepo           domainrepository.PinRepository
	messageRepo       domainrepository.MessageRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	userRepo          domainrepository.UserRepository
	notificationSvc   Notifier
	outputBuilder     *message.MessageOutputBuilder
	channelAccessSvc  service.ChannelAccessService
	systemMessageUC   systemmessage.UseCase
	permissionSvc     service.PermissionService
	searchIndexer     message.SearchIndexer
	logger            service.Logger
}

func NewPinInteractor(
	pinRepo domainrepository.PinRepository,
	messageRepo domainrepository.MessageRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc Notifier,
	outputBuilder *message.MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	systemMessageUC systemmessage.UseCase,
	permissionSvc service.PermissionService,
	searchIndexer message.SearchIndexer,
	logger service.Logger,
) PinUseCase {
	return &interactor{
		pinRepo:           pinRepo,
		messageRepo:       messageRepo,
		channelMemberRepo: channelMemberRepo,
		userRepo:          userRepo,
		notificationSvc:   notificationSvc,
		outputBuilder:     outputBuilder,
		channelAccessSvc:  channelAccessSvc,
		systemMessageUC:   systemMessageUC,
		permissionSvc:     permissionSvc,
		searchIndexer:     searchIndexer,
		logger:            logger,
	}
}

type PinMessageInput struct {
	ChannelID string
	MessageID string
	UserID    string
}

type UnpinMessageInput struct {
	ChannelID string
	MessageID string
	UserID    string
}

type ListPinsInput struct {
	ChannelID string
	UserID    string
	Limit     int
	Cursor    *string
}

type PinnedMessageOutput struct {
	Message  message.MessageOutput
	PinnedBy string
	PinnedAt time.Time
}

type ListPinsOutput struct {
	Pins       []PinnedMessageOutput
	NextCursor *string
}

// ensureCanPin はメッセージがチャンネルにあり、閲覧でき、ピン留めが許可されたロールであることを確認します
func (i *interactor) ensureCanPin(ctx context.Context, channelID, messageID, userID string) (*entity.Message, *entity.Channel, error) {
	msg, err := i.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg == nil || msg.ChannelID != channelID {
		return nil, nil, domerr.ErrMessageNotFound
	}
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := i.permissionSvc.Ensure(ctx, ch.WorkspaceID, userID, entity.PermissionPinMessages); err != nil {
		return nil, nil, err
	}
	return msg, ch, nil
}

func (i *interactor) PinMessage(ctx context.Context, input PinMessageInput) error {
	msg, ch, err := i.ensureCanPin(ctx, input.ChannelID, input.MessageID, input.UserID)
	if err != nil {
		return err
	}

	p := &entity.MessagePin{
		ChannelID: input.ChannelID,
		MessageID: input.MessageID,
		PinnedBy:  input.UserID,
	}
	if err := i.pinRepo.Create(ctx, p); err != nil {
		return err
	}
	i.searchIndexer.Sync(ctx, input.MessageID)

	payload := map[string]any{
		"messageId": input.MessageID,
		"pinnedBy":  input.UserID,
	}
	// スレッドの返信はスレッドを開いて表示するため、親メッセージも渡す
	if msg.ParentID != nil {
		payload["parentId"] = *msg.ParentID
	}
	if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
		Channel:   ch,
		Kind:      entity.SystemMessageKindMessagePinned,
		Payload:   payload,
		ActorID:   &input.UserID,
	}); err != nil {
		i.logger.Warn("ピン留めのシステムメッセージを作成できません", service.LogField{Key: "error", Value: err.Error()})
	}

	notification := PinNotification{MessageID: input.MessageID, PinnedBy: p.PinnedBy, PinnedAt: p.PinnedAt}
	if u, err := i.userRepo.FindByID(ctx, p.PinnedBy); err == nil && u != nil {
		notification.PinnedByUser = &message.UserInfo{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL, IsApp: u.IsApp}
	}
	i.notificationSvc.NotifyPinCreated(ch.WorkspaceID, input.ChannelID, i.memberIDs(ctx, input.ChannelID), notification)
	return nil
}

func (i *interactor) UnpinMessage(ctx context.Context, input UnpinMessageInput) error {
	_, ch, err := i.ensureCanPin(ctx, input.ChannelID, input.MessageID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.pinRepo.Delete(ctx, input.ChannelID, input.MessageID); err != nil {
		return err
	}
	i.searchIndexer.Sync(ctx, input.MessageID)
	i.notificationSvc.NotifyPinDeleted(ch.WorkspaceID, input.ChannelID, i.memberIDs(ctx, input.ChannelID), PinNotification{
		MessageID: input.MessageID,
		PinnedBy:  input.UserID,
		PinnedAt:  time.Now(),
	})
	return nil
}

// memberIDs はピンの件数を知らせるチャンネルの参加者です。取得できなくてもピン留めは成功させる
func (i *interactor) memberIDs(ctx context.Context, channelID string) []string {
	members, err := i.channelMemberRepo.FindMembers(ctx, channelID)
	if err != nil {
		i.logger.Warn("ピン留めを知らせる参加者を取得できません", service.LogField{Key: "error", Value: err.Error()})
		return nil
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids
}

func (i *interactor) ListPins(ctx context.Context, input ListPinsInput) (*ListPinsOutput, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 100
	}

	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}

	pins, next, err := i.pinRepo.List(ctx, input.ChannelID, input.Limit, input.Cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}

	pinned := make([]*entity.MessagePin, 0, len(pins))
	messages := make([]*entity.Message, 0, len(pins))
	for _, p := range pins {
		if p.Message != nil {
			pinned = append(pinned, p)
			messages = append(messages, p.Message)
		}
	}
	messageOutputs, err := i.outputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, err
	}

	outputs := make([]PinnedMessageOutput, 0, len(messageOutputs))
	for idx, msgOut := range messageOutputs {
		outputs = append(outputs, PinnedMessageOutput{
			Message:  msgOut,
			PinnedBy: pinned[idx].PinnedBy,
			PinnedAt: pinned[idx].PinnedAt,
		})
	}

	return &ListPinsOutput{Pins: outputs, NextCursor: next}, nil
}
