package pin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	domainservice "github.com/newt239/chat/internal/domain/service"
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
	notificationSvc   service.NotificationService
	messageAssembler  *message.MessageOutputAssembler
	channelAccessSvc  domainservice.ChannelAccessService
	systemMessageUC   systemmessage.UseCase
}

func NewPinInteractor(
	pinRepo domainrepository.PinRepository,
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	notificationSvc service.NotificationService,
	channelAccessSvc domainservice.ChannelAccessService,
	systemMessageUC systemmessage.UseCase,
) PinUseCase {
	return &interactor{
		pinRepo:           pinRepo,
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		notificationSvc:   notificationSvc,
		messageAssembler:  message.NewMessageOutputAssembler(),
		channelAccessSvc:  channelAccessSvc,
		systemMessageUC:   systemMessageUC,
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
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
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
		payload := map[string]interface{}{
			"message":  p.Message.ID,
			"pinnedBy": p.PinnedBy,
			"pinnedAt": p.PinnedAt.Format(time.RFC3339),
		}
		if workspaceID != "" {
			i.notificationSvc.NotifyPinCreated(workspaceID, input.ChannelID, payload)
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
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return err
	}

	if err := i.pinRepo.Delete(ctx, input.ChannelID, input.MessageID); err != nil {
		return err
	}
	if i.notificationSvc != nil {
		ch, _ := i.channelRepo.FindByID(ctx, input.ChannelID)
		if ch != nil {
			payload := map[string]interface{}{
				"message":  input.MessageID,
				"pinnedBy": input.UserID,
				"pinnedAt": time.Now().Format(time.RFC3339),
			}
			i.notificationSvc.NotifyPinDeleted(ch.WorkspaceID, input.ChannelID, payload)
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

	// メッセージのユーザー情報などを組み立て
	// すでに repository で Message を WithUser/WithChannel 済み -> Assembler で出力化
	// 補助データを最小にするため必要ユーザーだけ読み足し
	// Assembler は mentions/links/reactions/attachments も必要だが、
	// bookmark と同様の一覧を踏襲する場合は messageAssembler のみで十分

	// 収集用
	users := map[string]*entity.User{}
	userIDs := []string{}
	for _, p := range pins {
		if p.Message != nil && p.Message.UserID != "" {
			users[p.Message.UserID] = nil
		}
	}
	for id := range users {
		userIDs = append(userIDs, id)
	}
	if len(userIDs) > 0 {
		found, err := i.userRepo.FindByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load users: %w", err)
		}
		for _, u := range found {
			users[u.ID] = u
		}
	}

	outputs := make([]PinnedMessageOutput, 0, len(pins))
	for _, p := range pins {
		if p.Message == nil {
			continue
		}
		msgOut := i.messageAssembler.AssembleMessageOutput(
			p.Message,
			users[p.Message.UserID],
			nil, nil, nil, nil, nil,
			nil,
			users,
		)
		outputs = append(outputs, PinnedMessageOutput{
			Message:  msgOut,
			PinnedBy: p.PinnedBy,
			PinnedAt: p.PinnedAt,
		})
	}

	return &ListPinsOutput{Pins: outputs, NextCursor: next}, nil
}

// ensureChannelAccess は ChannelAccessService に委譲済み
