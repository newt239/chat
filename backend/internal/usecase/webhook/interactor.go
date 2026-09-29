package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/audit"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

// MaxTextLength は 1 回の投稿で受け付ける本文の最大文字数です
const MaxTextLength = 4000

const maxSenderNameLength = 80

var (
	ErrWebhookNotFound    = errors.New("指定された Webhook が見つかりません")
	ErrUnauthorized       = errors.New("この Webhook を編集できるのは発行者と管理者だけです")
	ErrUnsupportedChannel = errors.New("DM には Webhook を追加できません")
	ErrInactive           = errors.New("発行者がチャンネルを閲覧できなくなったため、この Webhook は使えません")
	ErrEmptyText          = fmt.Errorf("%w: text を指定してください", domerr.ErrValidation)
	ErrTextTooLong        = fmt.Errorf("%w: text は %d 文字以内で指定してください", domerr.ErrValidation, MaxTextLength)
	ErrInvalidAvatarURL   = fmt.Errorf("%w: avatar_url には http(s) の URL を指定してください", domerr.ErrValidation)
)

// MessagePoster はボットユーザー名義でメッセージを投稿して配信します
type MessagePoster interface {
	CreateBotMessage(ctx context.Context, channel *entity.Channel, message *entity.Message) (*messageuc.MessageOutput, error)
}

type Interactor struct {
	webhookRepo      domainrepository.WebhookRepository
	userRepo         domainrepository.UserRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	channelAccessSvc domainservice.ChannelAccessService
	poster           MessagePoster
	txManager        domaintransaction.Manager
	recorder         audit.Recorder
}

func NewInteractor(
	webhookRepo domainrepository.WebhookRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	poster MessagePoster,
	txManager domaintransaction.Manager,
	recorder audit.Recorder,
) *Interactor {
	return &Interactor{
		webhookRepo:      webhookRepo,
		userRepo:         userRepo,
		workspaceRepo:    workspaceRepo,
		channelAccessSvc: channelAccessSvc,
		poster:           poster,
		txManager:        txManager,
		recorder:         recorder,
	}
}

func (i *Interactor) List(ctx context.Context, input ListInput) ([]Output, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	webhooks, err := i.webhookRepo.FindByChannelID(ctx, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load webhooks: %w", err)
	}
	isAdmin, err := i.isAdmin(ctx, ch, input.UserID)
	if err != nil {
		return nil, err
	}
	return i.toOutputs(ctx, webhooks, input.UserID, isAdmin)
}

func (i *Interactor) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	if ch.Type == entity.ChannelTypeDM || ch.Type == entity.ChannelTypeGroupDM {
		return nil, ErrUnsupportedChannel
	}
	if ch.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}
	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	webhook := &entity.Webhook{
		ChannelID: ch.ID,
		Name:      strings.TrimSpace(input.Name),
		AvatarURL: input.AvatarURL,
		TokenHash: hash,
		CreatedBy: input.UserID,
	}
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		// パスワードは照合できない値にし、メールアドレスは配送されない .invalid ドメインにする
		bot := &entity.User{
			Email:        fmt.Sprintf("webhook-%s@webhook.invalid", uuid.NewString()),
			PasswordHash: entity.UnusablePasswordHash,
			DisplayName:  webhook.Name,
			AvatarURL:    webhook.AvatarURL,
			IsBot:        true,
		}
		if err := i.userRepo.Create(txCtx, bot); err != nil {
			return fmt.Errorf("failed to create bot user: %w", err)
		}
		webhook.BotUserID = bot.ID
		if err := i.webhookRepo.Create(txCtx, webhook); err != nil {
			return fmt.Errorf("failed to create webhook: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	i.record(ctx, ch, webhook, input.UserID, entity.AuditActionWebhookCreated)

	outputs, err := i.toOutputs(ctx, []*entity.Webhook{webhook}, input.UserID, false)
	if err != nil {
		return nil, err
	}
	return &CreateOutput{Webhook: outputs[0], Token: token}, nil
}

func (i *Interactor) Update(ctx context.Context, input UpdateInput) (*Output, error) {
	webhook, _, err := i.findManageable(ctx, input.WebhookID, input.UserID)
	if err != nil {
		return nil, err
	}
	webhook.Name = strings.TrimSpace(input.Name)
	webhook.AvatarURL = input.AvatarURL

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.webhookRepo.Update(txCtx, webhook); err != nil {
			return fmt.Errorf("failed to update webhook: %w", err)
		}
		// 過去の投稿の表示名も新しい名前にそろえる
		bot, err := i.userRepo.FindByID(txCtx, webhook.BotUserID)
		if err != nil {
			return fmt.Errorf("failed to load bot user: %w", err)
		}
		if bot == nil {
			return nil
		}
		bot.DisplayName = webhook.Name
		bot.AvatarURL = webhook.AvatarURL
		if err := i.userRepo.Update(txCtx, bot); err != nil {
			return fmt.Errorf("failed to update bot user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	outputs, err := i.toOutputs(ctx, []*entity.Webhook{webhook}, input.UserID, true)
	if err != nil {
		return nil, err
	}
	return &outputs[0], nil
}

func (i *Interactor) RegenerateToken(ctx context.Context, input TargetInput) (string, error) {
	webhook, _, err := i.findManageable(ctx, input.WebhookID, input.UserID)
	if err != nil {
		return "", err
	}
	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	webhook.TokenHash = hash
	if err := i.webhookRepo.Update(ctx, webhook); err != nil {
		return "", fmt.Errorf("failed to update webhook: %w", err)
	}
	return token, nil
}

func (i *Interactor) Delete(ctx context.Context, input TargetInput) error {
	webhook, ch, err := i.findManageable(ctx, input.WebhookID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.webhookRepo.Delete(ctx, webhook.ID); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	i.record(ctx, ch, webhook, input.UserID, entity.AuditActionWebhookDeleted)
	return nil
}

// Post は外部から届いた内容を Webhook のボットユーザー名義でチャンネルに投稿します
func (i *Interactor) Post(ctx context.Context, input PostInput) (*messageuc.MessageOutput, error) {
	if uuid.Validate(input.WebhookID) != nil {
		return nil, ErrWebhookNotFound
	}
	webhook, err := i.webhookRepo.FindByID(ctx, input.WebhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to load webhook: %w", err)
	}
	// トークンの誤りも存在しない場合と区別しない
	if webhook == nil || !webhook.VerifyToken(input.Token) {
		return nil, ErrWebhookNotFound
	}

	message, err := buildMessage(webhook, input)
	if err != nil {
		return nil, err
	}

	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, webhook.ChannelID, webhook.CreatedBy)
	if errors.Is(err, domerr.ErrUnauthorized) {
		return nil, ErrInactive
	}
	if err != nil {
		return nil, err
	}
	if ch.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}

	output, err := i.poster.CreateBotMessage(ctx, ch, message)
	if err != nil {
		return nil, err
	}
	if err := i.webhookRepo.MarkUsed(ctx, webhook.ID, output.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to update last used time: %w", err)
	}
	return output, nil
}

func buildMessage(webhook *entity.Webhook, input PostInput) (*entity.Message, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return nil, ErrEmptyText
	}
	if utf8.RuneCountInString(text) > MaxTextLength {
		return nil, ErrTextTooLong
	}

	message := &entity.Message{UserID: webhook.BotUserID, Body: text}
	if input.Username != nil {
		if name := truncate(strings.TrimSpace(*input.Username), maxSenderNameLength); name != "" {
			message.SenderName = &name
		}
	}
	if input.AvatarURL != nil && strings.TrimSpace(*input.AvatarURL) != "" {
		avatar := strings.TrimSpace(*input.AvatarURL)
		if !isHTTPURL(avatar) {
			return nil, ErrInvalidAvatarURL
		}
		message.SenderAvatarURL = &avatar
	}
	return message, nil
}

// findManageable は Webhook を取得し、発行者か管理者であることを確認します
func (i *Interactor) findManageable(ctx context.Context, webhookID, userID string) (*entity.Webhook, *entity.Channel, error) {
	webhook, err := i.webhookRepo.FindByID(ctx, webhookID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load webhook: %w", err)
	}
	if webhook == nil {
		return nil, nil, ErrWebhookNotFound
	}
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, webhook.ChannelID, userID)
	if err != nil {
		return nil, nil, err
	}
	if webhook.CreatedBy == userID {
		return webhook, ch, nil
	}
	isAdmin, err := i.isAdmin(ctx, ch, userID)
	if err != nil {
		return nil, nil, err
	}
	if !isAdmin {
		return nil, nil, ErrUnauthorized
	}
	return webhook, ch, nil
}

func (i *Interactor) isAdmin(ctx context.Context, ch *entity.Channel, userID string) (bool, error) {
	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	return member.IsAdmin(), nil
}

func (i *Interactor) toOutputs(ctx context.Context, webhooks []*entity.Webhook, viewerID string, isAdmin bool) ([]Output, error) {
	creatorIDs := make([]string, 0, len(webhooks))
	for _, w := range webhooks {
		creatorIDs = append(creatorIDs, w.CreatedBy)
	}
	creators, err := i.userRepo.FindByIDs(ctx, creatorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load creators: %w", err)
	}
	byID := make(map[string]*entity.User, len(creators))
	for _, u := range creators {
		byID[u.ID] = u
	}

	outputs := make([]Output, 0, len(webhooks))
	for _, w := range webhooks {
		creator := messageuc.UserInfo{ID: w.CreatedBy}
		if u := byID[w.CreatedBy]; u != nil {
			creator = messageuc.UserInfo{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL}
		}
		outputs = append(outputs, Output{
			ID:         w.ID,
			ChannelID:  w.ChannelID,
			Name:       w.Name,
			AvatarURL:  w.AvatarURL,
			CreatedBy:  creator,
			CreatedAt:  w.CreatedAt,
			LastUsedAt: w.LastUsedAt,
			CanManage:  isAdmin || w.CreatedBy == viewerID,
		})
	}
	return outputs, nil
}

func (i *Interactor) record(ctx context.Context, ch *entity.Channel, webhook *entity.Webhook, actorID string, action entity.AuditAction) {
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: ch.WorkspaceID,
		ActorID:     &actorID,
		Action:      action,
		TargetType:  entity.AuditTargetWebhook,
		TargetID:    webhook.ID,
		TargetLabel: webhook.Name,
		Metadata:    map[string]string{"channel": ch.Name},
	})
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}
