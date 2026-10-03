package pin

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

const maxPins = 100

type PinInput struct {
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

type Interactor struct {
	pinRepo           domainrepository.PinRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	userRepo          domainrepository.UserRepository
	notificationSvc   Notifier
	outputBuilder     *message.MessageOutputBuilder
	channelAccessSvc  service.ChannelAccessService
	systemMessages    *message.SystemMessages
	permissionSvc     service.PermissionService
	searchIndexer     message.SearchIndexer
}

func New(
	pinRepo domainrepository.PinRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc Notifier,
	outputBuilder *message.MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	systemMessages *message.SystemMessages,
	permissionSvc service.PermissionService,
	searchIndexer message.SearchIndexer,
) *Interactor {
	return &Interactor{
		pinRepo:           pinRepo,
		channelMemberRepo: channelMemberRepo,
		userRepo:          userRepo,
		notificationSvc:   notificationSvc,
		outputBuilder:     outputBuilder,
		channelAccessSvc:  channelAccessSvc,
		systemMessages:    systemMessages,
		permissionSvc:     permissionSvc,
		searchIndexer:     searchIndexer,
	}
}

// ensureCanPin はメッセージがチャンネルにあり、閲覧でき、ピン留めが許可されたロールであることを確認します
func (i *Interactor) ensureCanPin(ctx context.Context, input PinInput) (*entity.Message, *entity.Channel, error) {
	msg, ch, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return nil, nil, err
	}
	if msg.ChannelID != input.ChannelID {
		return nil, nil, domerr.ErrMessageNotFound
	}
	if _, err := i.permissionSvc.Ensure(ctx, ch.WorkspaceID, input.UserID, entity.PermissionPinMessages); err != nil {
		return nil, nil, err
	}
	return msg, ch, nil
}

func (i *Interactor) PinMessage(ctx context.Context, input PinInput) error {
	msg, ch, err := i.ensureCanPin(ctx, input)
	if err != nil {
		return err
	}
	p := &entity.MessagePin{ChannelID: input.ChannelID, MessageID: input.MessageID, PinnedBy: input.UserID}
	if err := i.pinRepo.Create(ctx, p); err != nil {
		return err
	}
	i.searchIndexer.Sync(ctx, input.MessageID)

	payload := map[string]any{"messageId": input.MessageID, "pinnedBy": input.UserID}
	// スレッドの返信はスレッドを開いて表示するため、親メッセージも渡す
	if msg.ParentID != nil {
		payload["parentId"] = *msg.ParentID
	}
	i.systemMessages.Record(ctx, ch, entity.SystemMessageKindMessagePinned, input.UserID, payload)

	notification := PinNotification{MessageID: input.MessageID, PinnedBy: p.PinnedBy, PinnedAt: p.PinnedAt}
	if u, err := i.userRepo.FindByID(ctx, p.PinnedBy); err == nil && u != nil {
		notification.PinnedByUser = new(message.NewUserInfo(u))
	}
	i.notificationSvc.NotifyPinCreated(ch.WorkspaceID, input.ChannelID, i.memberIDs(ctx, input.ChannelID), notification)
	return nil
}

func (i *Interactor) UnpinMessage(ctx context.Context, input PinInput) error {
	_, ch, err := i.ensureCanPin(ctx, input)
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
func (i *Interactor) memberIDs(ctx context.Context, channelID string) []string {
	members, err := i.channelMemberRepo.FindMembers(ctx, channelID)
	if err != nil {
		slog.WarnContext(ctx, "ピン留めを知らせる参加者を取得できません", "error", err)
		return nil
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids
}

func (i *Interactor) ListPins(ctx context.Context, input ListPinsInput) (*ListPinsOutput, error) {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}
	pins, next, err := i.pinRepo.List(ctx, input.ChannelID, min(cmp.Or(input.Limit, maxPins), maxPins), input.Cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}
	messages := make([]*entity.Message, len(pins))
	for idx, p := range pins {
		messages[idx] = p.Message
	}
	messageOutputs, err := i.outputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, err
	}
	outputs := make([]PinnedMessageOutput, len(pins))
	for idx, p := range pins {
		outputs[idx] = PinnedMessageOutput{Message: messageOutputs[idx], PinnedBy: p.PinnedBy, PinnedAt: p.PinnedAt}
	}
	return &ListPinsOutput{Pins: outputs, NextCursor: next}, nil
}
