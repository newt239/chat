package scheduledmessage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrScheduledMessageNotFound = errors.New("予約メッセージが見つかりません")
	ErrScheduleInPast           = fmt.Errorf("%w: 予約日時は現在より後にしてください", domerr.ErrValidation)
	ErrNotEditable              = errors.New("送信中または送信済みの予約は変更できません")
	errSendFailed               = errors.New("送信に失敗しました")
)

// 送信の失敗理由としてそのまま利用者に見せるエラー。それ以外は内部の詳細を隠す
var userFacingErrors = []error{
	domerr.ErrUnauthorized, domerr.ErrChannelNotFound, domerr.ErrChannelArchived,
	domerr.ErrParentMessageNotFound, domerr.ErrAttachmentNotFound, messageuc.ErrEmptyMessage,
}

const (
	dispatchBatchSize = 50
	// 送信中のまま staleSendingAfter を過ぎた予約は、送信の途中でサーバーが止まったものとして失敗にする
	staleSendingAfter = 5 * time.Minute
	// 期限切れのセッションを消す間隔
	sessionCleanupInterval = time.Hour
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
	sessionRepo      domainrepository.SessionRepository
	channelAccessSvc service.ChannelAccessService
	poster           MessagePoster
	logger           service.Logger
}

func NewInteractor(
	scheduledRepo domainrepository.ScheduledMessageRepository,
	messageRepo domainrepository.MessageRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	sessionRepo domainrepository.SessionRepository,
	channelAccessSvc service.ChannelAccessService,
	poster MessagePoster,
	logger service.Logger,
) *Interactor {
	return &Interactor{
		scheduledRepo:    scheduledRepo,
		messageRepo:      messageRepo,
		attachmentRepo:   attachmentRepo,
		sessionRepo:      sessionRepo,
		channelAccessSvc: channelAccessSvc,
		poster:           poster,
		logger:           logger,
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
	channel, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	if channel.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}
	if err := i.ensureParent(ctx, input.ParentID, input.ChannelID); err != nil {
		return nil, err
	}
	if err := i.ensureAttachments(ctx, input); err != nil {
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

// RunDispatcher は ctx が終わるまで interval ごとに期限の来た予約を送り、ついでに期限切れのセッションを消します
func (i *Interactor) RunDispatcher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var cleanedAt time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if _, err := i.DispatchDue(ctx); err != nil {
				i.logger.Error("予約メッセージの送信処理に失敗しました", service.LogField{Key: "error", Value: err.Error()})
			}
			if now.Sub(cleanedAt) < sessionCleanupInterval {
				continue
			}
			cleanedAt = now
			if err := i.sessionRepo.DeleteExpired(ctx); err != nil {
				i.logger.Error("期限切れのセッションを削除できません", service.LogField{Key: "error", Value: err.Error()})
			}
		}
	}
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
		err = i.scheduledRepo.MarkSent(ctx, message.ID, out.ID)
		if err != nil {
			i.logger.Error("予約メッセージの送信済みを記録できません", service.LogField{Key: "id", Value: message.ID}, service.LogField{Key: "error", Value: err.Error()})
		}
		return
	}
	reason := failureReason(err)
	if reason == errSendFailed.Error() {
		i.logger.Error("予約メッセージを送信できません", service.LogField{Key: "id", Value: message.ID}, service.LogField{Key: "error", Value: err.Error()})
	}
	if err := i.scheduledRepo.MarkFailed(ctx, message.ID, reason); err != nil {
		i.logger.Error("予約メッセージの失敗を記録できません", service.LogField{Key: "id", Value: message.ID}, service.LogField{Key: "error", Value: err.Error()})
	}
}

func failureReason(err error) string {
	for _, target := range userFacingErrors {
		if errors.Is(err, target) {
			return target.Error()
		}
	}
	return errSendFailed.Error()
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

func (i *Interactor) ensureParent(ctx context.Context, parentID *string, channelID string) error {
	if parentID == nil {
		return nil
	}
	parent, err := i.messageRepo.FindByID(ctx, *parentID)
	if err != nil {
		return fmt.Errorf("failed to load parent message: %w", err)
	}
	if !parent.CanBeRepliedIn(channelID) {
		return domerr.ErrParentMessageNotFound
	}
	return nil
}

// ensureAttachments は添付が本人のまだ使っていないもので、同じチャンネル宛てかを確かめます
func (i *Interactor) ensureAttachments(ctx context.Context, input ScheduleInput) error {
	if len(input.AttachmentIDs) == 0 {
		return nil
	}
	attachments, err := i.attachmentRepo.FindPendingByIDsForUser(ctx, input.UserID, input.AttachmentIDs)
	if err != nil {
		return fmt.Errorf("failed to verify attachments: %w", err)
	}
	if len(attachments) != len(input.AttachmentIDs) {
		return domerr.ErrAttachmentNotFound
	}
	for _, attachment := range attachments {
		if attachment.ChannelID != input.ChannelID {
			return domerr.ErrAttachmentNotFound
		}
	}
	return nil
}
