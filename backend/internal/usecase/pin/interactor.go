package pin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

var (
	ErrUnauthorized    = errors.New("この操作を行う権限がありません")
	ErrMessageNotFound = errors.New("メッセージが見つかりません")
	ErrPinExists       = errors.New("このメッセージは既にピン留めされています")
)

type PinUseCase interface {
	PinMessage(ctx context.Context, input PinMessageInput) error
	UnpinMessage(ctx context.Context, input UnpinMessageInput) error
	ListPins(ctx context.Context, input ListPinsInput) (*ListPinsOutput, error)
}

type interactor struct {
	pinRepo           domainrepository.PinRepository
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
	notificationSvc   Notifier
	outputBuilder     *message.MessageOutputBuilder
	channelAccessSvc  service.ChannelAccessService
	systemMessageUC   systemmessage.UseCase
	permissionSvc     service.PermissionService
	searchIndexer     message.SearchIndexer
}

func NewPinInteractor(
	pinRepo domainrepository.PinRepository,
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc Notifier,
	outputBuilder *message.MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	systemMessageUC systemmessage.UseCase,
	permissionSvc service.PermissionService,
	searchIndexer message.SearchIndexer,
) PinUseCase {
	return &interactor{
		pinRepo:           pinRepo,
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		notificationSvc:   notificationSvc,
		outputBuilder:     outputBuilder,
		channelAccessSvc:  channelAccessSvc,
		systemMessageUC:   systemMessageUC,
		permissionSvc:     permissionSvc,
		searchIndexer:     searchIndexer,
	}
}

// ensureCanPin はチャンネルを閲覧でき、ピン留めが許可されたロールであることを確認します
func (i *interactor) ensureCanPin(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return err
	}
	_, err = i.permissionSvc.Ensure(ctx, ch.WorkspaceID, userID, entity.PermissionPinMessages)
	return err
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

func (i *interactor) PinMessage(ctx context.Context, input PinMessageInput) error {
	// メッセージ存在確認
	msg, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg == nil || msg.ChannelID != input.ChannelID {
		return ErrMessageNotFound
	}

	// アクセス権確認
	if err := i.ensureCanPin(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}

	// 作成（ユニーク制約違反はリポジトリ側でDBエラーになるが、409として扱いたいのでここではそのまま返す）
	p := &entity.MessagePin{
		ChannelID: input.ChannelID,
		MessageID: input.MessageID,
		PinnedBy:  input.UserID,
		PinnedAt:  time.Now(),
	}
	if err := i.pinRepo.Create(ctx, p); err != nil {
		return err
	}
	i.searchIndexer.Sync(ctx, input.MessageID)

	// システムメッセージ作成（ピン留め）
	if i.systemMessageUC != nil {
		payload := map[string]any{
			"messageId": input.MessageID,
			"pinnedBy":  input.UserID,
		}
		actorID := input.UserID
		_, _ = i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
			ChannelID: input.ChannelID,
			Kind:      entity.SystemMessageKindMessagePinned,
			Payload:   payload,
			ActorID:   &actorID,
		})
	}
	// 通知
	if i.notificationSvc != nil && p.Message != nil {
		workspaceID := ""
		if p.Message.ChannelID != "" {
			// チャンネルのワークスペースID取得のために再取得
			ch, _ := i.channelRepo.FindByID(ctx, p.Message.ChannelID)
			if ch != nil {
				workspaceID = ch.WorkspaceID
			}
		}
		if workspaceID != "" {
			notification := PinNotification{MessageID: p.Message.ID, PinnedBy: p.PinnedBy, PinnedAt: p.PinnedAt}
			if u, _ := i.userRepo.FindByID(ctx, p.PinnedBy); u != nil {
				notification.PinnedByUser = &message.UserInfo{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL}
			}
			i.notificationSvc.NotifyPinCreated(workspaceID, input.ChannelID, notification)
		}
	}
	return nil
}

func (i *interactor) UnpinMessage(ctx context.Context, input UnpinMessageInput) error {
	// メッセージ存在確認
	msg, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg == nil || msg.ChannelID != input.ChannelID {
		return ErrMessageNotFound
	}

	// アクセス権確認
	if err := i.ensureCanPin(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}

	if err := i.pinRepo.Delete(ctx, input.ChannelID, input.MessageID); err != nil {
		return err
	}
	i.searchIndexer.Sync(ctx, input.MessageID)
	if i.notificationSvc != nil {
		ch, _ := i.channelRepo.FindByID(ctx, input.ChannelID)
		if ch != nil {
			i.notificationSvc.NotifyPinDeleted(ch.WorkspaceID, input.ChannelID, PinNotification{
				MessageID: input.MessageID,
				PinnedBy:  input.UserID,
				PinnedAt:  time.Now(),
			})
		}
	}
	return nil
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

// ensureChannelAccess は ChannelAccessService に委譲済み
