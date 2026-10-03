package scheduledmessage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrScheduledMessageNotFound = domerr.New(domerr.ErrNotFound, "予約メッセージが見つかりません")
	ErrScheduleInPast           = domerr.New(domerr.ErrValidation, "予約日時は現在より後にしてください")
	ErrNotEditable              = domerr.New(domerr.ErrFailedPrecondition, "送信中または送信済みの予約は変更できません")
)

const (
	dispatchBatchSize = 50
	// 送信中のまま staleSendingAfter を過ぎた予約は、送信の途中でサーバーが止まったものとして失敗にする
	staleSendingAfter = 5 * time.Minute
)

// MessagePoster は予約した内容を通常の投稿と同じ経路で投稿します
type MessagePoster interface {
	CreateMessage(ctx context.Context, input messageuc.CreateMessageInput) (*messageuc.MessageOutput, error)
}

type ScheduleInput struct {
	UserID        string
	ChannelID     string
	ParentID      *string
	Body          string
	AttachmentIDs []string
	Location      *entity.MessageLocation
	ScheduledAt   time.Time
}

type RescheduleInput struct {
	ID          string
	UserID      string
	Body        string
	ScheduledAt time.Time
}

type Interactor struct {
	scheduledRepo    domainrepository.ScheduledMessageRepository
	messageRepo      domainrepository.MessageRepository
	attachmentRepo   domainrepository.AttachmentRepository
	channelAccessSvc service.ChannelAccessService
	poster           MessagePoster
}

func New(
	scheduledRepo domainrepository.ScheduledMessageRepository,
	messageRepo domainrepository.MessageRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	channelAccessSvc service.ChannelAccessService,
	poster MessagePoster,
) *Interactor {
	return &Interactor{
		scheduledRepo:    scheduledRepo,
		messageRepo:      messageRepo,
		attachmentRepo:   attachmentRepo,
		channelAccessSvc: channelAccessSvc,
		poster:           poster,
	}
}

// Schedule は投稿できる内容かを今の時点で確かめてから予約します。送信時にも改めて確かめる
func (i *Interactor) Schedule(ctx context.Context, input ScheduleInput) (*entity.ScheduledMessage, error) {
	if strings.TrimSpace(input.Body) == "" && len(input.AttachmentIDs) == 0 && input.Location == nil {
		return nil, messageuc.ErrEmptyMessage
	}
	if !input.ScheduledAt.After(time.Now()) {
		return nil, ErrScheduleInPast
	}
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
	}
	if _, err := messageuc.EnsureReplyTarget(ctx, i.messageRepo, input.ParentID, input.ChannelID); err != nil {
		return nil, err
	}
	if err := messageuc.VerifyAttachments(ctx, i.attachmentRepo, input.UserID, input.ChannelID, input.AttachmentIDs); err != nil {
		return nil, err
	}

	message := &entity.ScheduledMessage{
		UserID:        input.UserID,
		ChannelID:     input.ChannelID,
		ParentID:      input.ParentID,
		Body:          input.Body,
		AttachmentIDs: input.AttachmentIDs,
		Location:      input.Location,
		ScheduledAt:   input.ScheduledAt,
	}
	if err := i.scheduledRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create scheduled message: %w", err)
	}
	return message, nil
}

func (i *Interactor) List(ctx context.Context, userID, workspaceID string) ([]*entity.ScheduledMessage, error) {
	return i.scheduledRepo.FindByWorkspace(ctx, userID, workspaceID)
}

// Reschedule は本文と日時を変えます。失敗した予約はこれで予約中に戻る
func (i *Interactor) Reschedule(ctx context.Context, input RescheduleInput) (*entity.ScheduledMessage, error) {
	message, err := i.findOwn(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}
	if !message.IsEditable() {
		return nil, ErrNotEditable
	}
	if strings.TrimSpace(input.Body) == "" && len(message.AttachmentIDs) == 0 && message.Location == nil {
		return nil, messageuc.ErrEmptyMessage
	}
	if !input.ScheduledAt.After(time.Now()) {
		return nil, ErrScheduleInPast
	}
	if err := i.scheduledRepo.Reschedule(ctx, input.ID, input.Body, input.ScheduledAt); err != nil {
		return nil, fmt.Errorf("failed to reschedule: %w", err)
	}
	return i.scheduledRepo.FindByID(ctx, input.ID)
}

func (i *Interactor) Delete(ctx context.Context, id, userID string) error {
	message, err := i.findOwn(ctx, id, userID)
	if err != nil {
		return err
	}
	if message.Status == entity.ScheduledMessageSending {
		return ErrNotEditable
	}
	return i.scheduledRepo.Delete(ctx, id)
}

// SendNow は予約日時を待たずに投稿し、結果を反映した予約を返します
func (i *Interactor) SendNow(ctx context.Context, id, userID string) (*entity.ScheduledMessage, error) {
	if _, err := i.findOwn(ctx, id, userID); err != nil {
		return nil, err
	}
	message, err := i.scheduledRepo.Claim(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to claim scheduled message: %w", err)
	}
	if message == nil {
		return nil, ErrNotEditable
	}
	i.send(ctx, message)
	return i.scheduledRepo.FindByID(ctx, id)
}

// DispatchDue は期限の来た予約を送り、処理した件数を返します
func (i *Interactor) DispatchDue(ctx context.Context) (int, error) {
	now := time.Now()
	messages, err := i.scheduledRepo.ClaimDue(ctx, now, now.Add(-staleSendingAfter), dispatchBatchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to claim due messages: %w", err)
	}
	for _, message := range messages {
		i.send(ctx, message)
	}
	return len(messages), nil
}

// send は通常の投稿と同じ経路（権限の確認・メンション・検索インデックス・配信）で投稿し、結果を記録します
func (i *Interactor) send(ctx context.Context, message *entity.ScheduledMessage) {
	out, err := i.poster.CreateMessage(ctx, messageuc.CreateMessageInput{
		ChannelID:     message.ChannelID,
		UserID:        message.UserID,
		Body:          message.Body,
		ParentID:      message.ParentID,
		AttachmentIDs: message.AttachmentIDs,
		Location:      message.Location,
	})
	if err == nil {
		if err := i.scheduledRepo.MarkSent(ctx, message.ID, out.ID); err != nil {
			slog.ErrorContext(ctx, "予約メッセージの送信済みを記録できません", "id", message.ID, "error", err)
		}
		return
	}
	if err := i.scheduledRepo.MarkFailed(ctx, message.ID, failureReason(ctx, message.ID, err)); err != nil {
		slog.ErrorContext(ctx, "予約メッセージの失敗を記録できません", "id", message.ID, "error", err)
	}
}

// failureReason は権限・存在・入力の誤りならそのまま利用者に見せ、それ以外は内部の詳細を隠します
func failureReason(ctx context.Context, id string, err error) string {
	for _, kind := range []error{domerr.ErrUnauthorized, domerr.ErrNotFound, domerr.ErrValidation} {
		if errors.Is(err, kind) {
			return err.Error()
		}
	}
	slog.ErrorContext(ctx, "予約メッセージを送信できません", "id", id, "error", err)
	return "送信に失敗しました"
}

// findOwn は他人の予約を存在しないものとして扱います
func (i *Interactor) findOwn(ctx context.Context, id, userID string) (*entity.ScheduledMessage, error) {
	message, err := i.scheduledRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load scheduled message: %w", err)
	}
	if message == nil || message.UserID != userID {
		return nil, ErrScheduledMessageNotFound
	}
	return message, nil
}
