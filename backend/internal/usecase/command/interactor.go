// Package command はメッセージの入力欄から実行するスラッシュコマンドを扱います。応答は公式アプリが投稿する
package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var ErrUnknownCommand = fmt.Errorf("%w: 不明なコマンドです", domerr.ErrValidation)

const dispatchBatchSize = 50

// OfficialPoster は公式アプリの名義で投稿します
type OfficialPoster interface {
	EnsureOfficial(ctx context.Context, workspaceID string) (*entity.App, error)
	PostAsOfficial(ctx context.Context, workspaceID, channelID string, parentID *string, body string) (*messageuc.MessageOutput, error)
}

type ExecuteInput struct {
	UserID    string
	ChannelID string
	ParentID  *string
	// 「/」から始まる入力全体
	Text string
}

type Interactor struct {
	reminderRepo      domainrepository.ReminderRepository
	userRepo          domainrepository.UserRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	channelAccessSvc  domainservice.ChannelAccessService
	poster            OfficialPoster
	logger            domainservice.Logger
	now               func() time.Time
}

func NewInteractor(
	reminderRepo domainrepository.ReminderRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	poster OfficialPoster,
	logger domainservice.Logger,
) *Interactor {
	return &Interactor{
		reminderRepo:      reminderRepo,
		userRepo:          userRepo,
		workspaceRepo:     workspaceRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		channelAccessSvc:  channelAccessSvc,
		poster:            poster,
		logger:            logger,
		now:               time.Now,
	}
}

// Execute はコマンドを実行し、公式アプリがチャンネルに投稿した応答を返します
func (i *Interactor) Execute(ctx context.Context, input ExecuteInput) (*messageuc.MessageOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	if ch.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}
	name, args, _ := strings.Cut(strings.TrimPrefix(strings.TrimSpace(input.Text), "/"), " ")
	switch name {
	case "remind":
		return i.remind(ctx, ch, input, args)
	default:
		return nil, ErrUnknownCommand
	}
}

func (i *Interactor) remind(ctx context.Context, ch *entity.Channel, input ExecuteInput, args string) (*messageuc.MessageOutput, error) {
	loc, err := i.userLocation(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	req, err := ParseRemind(args, i.now().In(loc))
	if err != nil {
		return nil, err
	}
	if err := i.ensureTarget(ctx, ch.WorkspaceID, input.UserID, req.Target); err != nil {
		return nil, err
	}
	if err := i.reminderRepo.Create(ctx, &entity.Reminder{
		WorkspaceID:     ch.WorkspaceID,
		CreatorID:       input.UserID,
		TargetUserID:    req.Target.UserID,
		TargetChannelID: req.Target.ChannelID,
		Text:            req.Text,
		RemindAt:        req.At,
	}); err != nil {
		return nil, fmt.Errorf("failed to create reminder: %w", err)
	}

	when := req.At.Format("2006/01/02 15:04 MST")
	var body string
	switch {
	case req.Target.ChannelID != nil:
		body = fmt.Sprintf("<#%s> に %s にリマインドします: 「%s」（設定: <@%s>）", *req.Target.ChannelID, when, req.Text, input.UserID)
	case req.Target.UserID != nil && *req.Target.UserID != input.UserID:
		body = fmt.Sprintf("<@%s> さんに %s にリマインドします: 「%s」（設定: <@%s>）", *req.Target.UserID, when, req.Text, input.UserID)
	default:
		body = fmt.Sprintf("<@%s> さんに %s にリマインドします: 「%s」", input.UserID, when, req.Text)
	}
	return i.poster.PostAsOfficial(ctx, ch.WorkspaceID, ch.ID, input.ParentID, body)
}

// ensureTarget は届け先が同じワークスペースのメンバーか、実行した人が参加しているチャンネルであることを確かめます
func (i *Interactor) ensureTarget(ctx context.Context, workspaceID, userID string, target RemindTarget) error {
	if target.ChannelID != nil {
		ch, err := i.channelAccessSvc.EnsureChannelMember(ctx, *target.ChannelID, userID)
		if err != nil {
			return err
		}
		if ch.WorkspaceID != workspaceID {
			return domerr.ErrChannelNotFound
		}
		return nil
	}
	if target.UserID != nil {
		member, err := i.workspaceRepo.FindMember(ctx, workspaceID, *target.UserID)
		if err != nil {
			return fmt.Errorf("failed to verify target: %w", err)
		}
		if member == nil {
			return domerr.ErrNotFound
		}
	}
	return nil
}

// userLocation はプロフィールのタイムゾーンを返します。未設定なら UTC
func (i *Interactor) userLocation(ctx context.Context, userID string) (*time.Location, error) {
	user, err := i.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil || user.Preferences.Timezone == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(user.Preferences.Timezone)
	if err != nil {
		return time.UTC, nil
	}
	return loc, nil
}

// DispatchDue は期限の来たリマインダーを公式アプリから届け、処理した件数を返します
func (i *Interactor) DispatchDue(ctx context.Context) (int, error) {
	reminders, err := i.reminderRepo.ClaimDue(ctx, i.now(), dispatchBatchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to claim due reminders: %w", err)
	}
	for _, reminder := range reminders {
		if err := i.deliver(ctx, reminder); err != nil {
			i.logger.Error("リマインダーを届けられません", domainservice.LogField{Key: "id", Value: reminder.ID}, domainservice.LogField{Key: "error", Value: err.Error()})
			if err := i.reminderRepo.MarkFailed(ctx, reminder.ID); err != nil {
				i.logger.Error("リマインダーの失敗を記録できません", domainservice.LogField{Key: "id", Value: reminder.ID}, domainservice.LogField{Key: "error", Value: err.Error()})
			}
			continue
		}
		if err := i.reminderRepo.MarkSent(ctx, reminder.ID); err != nil {
			i.logger.Error("リマインダーの送信済みを記録できません", domainservice.LogField{Key: "id", Value: reminder.ID}, domainservice.LogField{Key: "error", Value: err.Error()})
		}
	}
	return len(reminders), nil
}

// RunDispatcher は ctx が終わるまで interval ごとに期限の来たリマインダーを届けます
func (i *Interactor) RunDispatcher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := i.DispatchDue(ctx); err != nil {
				i.logger.Error("リマインダーの送信処理に失敗しました", domainservice.LogField{Key: "error", Value: err.Error()})
			}
		}
	}
}

// deliver はチャンネル宛てならそのチャンネルに、人宛てなら公式アプリとの DM に投稿します
func (i *Interactor) deliver(ctx context.Context, reminder *entity.Reminder) error {
	if reminder.TargetChannelID != nil {
		body := fmt.Sprintf("<@%s> からのリマインダー: %s", reminder.CreatorID, reminder.Text)
		_, err := i.poster.PostAsOfficial(ctx, reminder.WorkspaceID, *reminder.TargetChannelID, nil, body)
		return err
	}
	recipient := reminder.CreatorID
	if reminder.TargetUserID != nil {
		recipient = *reminder.TargetUserID
	}
	official, err := i.poster.EnsureOfficial(ctx, reminder.WorkspaceID)
	if err != nil {
		return err
	}
	dm, err := i.channelRepo.FindOrCreateDM(ctx, reminder.WorkspaceID, official.BotUserID, recipient)
	if err != nil {
		return fmt.Errorf("failed to open DM: %w", err)
	}
	// 新しく作った DM にはメンバーがいないため、公式アプリと受け取る人を参加させる。参加済みなら何もしない
	for _, userID := range []string{official.BotUserID, recipient} {
		if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: dm.ID, UserID: userID, Role: entity.ChannelRoleMember, JoinedAt: time.Now()}); err != nil {
			return fmt.Errorf("failed to join DM: %w", err)
		}
	}
	body := fmt.Sprintf("<@%s> リマインダー: %s", recipient, reminder.Text)
	if recipient != reminder.CreatorID {
		body += fmt.Sprintf("（<@%s> から）", reminder.CreatorID)
	}
	_, err = i.poster.PostAsOfficial(ctx, reminder.WorkspaceID, dm.ID, nil, body)
	return err
}
