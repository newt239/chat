package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
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

// OfficialAppName は公式アプリの名前です
const OfficialAppName = "Chat"

var (
	ErrAppNotFound        = errors.New("指定されたアプリが見つかりません")
	ErrOfficialApp        = errors.New("公式アプリは編集・削除できません")
	ErrUnsupportedChannel = errors.New("DM にはアプリを追加できません")
	ErrInactive           = errors.New("作成者がワークスペースを抜けたため、このアプリは使えません")
	ErrForbiddenChannel   = errors.New("このアプリにはこのチャンネルへ投稿する権限がありません")
	ErrForbiddenThread    = errors.New("このアプリにはスレッドへ返信する権限がありません")
	ErrChannelRequired    = fmt.Errorf("%w: channel_id を指定するか、アプリの既定のチャンネルを設定してください", domerr.ErrValidation)
	ErrEmptyText          = fmt.Errorf("%w: text を指定してください", domerr.ErrValidation)
	ErrTextTooLong        = fmt.Errorf("%w: text は %d 文字以内で指定してください", domerr.ErrValidation, MaxTextLength)
	ErrInvalidURL         = fmt.Errorf("%w: http(s) の URL を指定してください", domerr.ErrValidation)
	ErrUnknownPermission  = fmt.Errorf("%w: 不明な権限です", domerr.ErrValidation)
	ErrOutgoingURLMissing = fmt.Errorf("%w: 送信 Webhook を許可するときは送信先の URL を指定してください", domerr.ErrValidation)
)

// MessagePoster はボットユーザー名義でメッセージを投稿して配信します
type MessagePoster interface {
	CreateBotMessage(ctx context.Context, channel *entity.Channel, message *entity.Message) (*messageuc.MessageOutput, error)
}

type Interactor struct {
	appRepo           domainrepository.AppRepository
	userRepo          domainrepository.UserRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	messageRepo       domainrepository.MessageRepository
	channelAccessSvc  domainservice.ChannelAccessService
	poster            MessagePoster
	txManager         domaintransaction.Manager
	recorder          audit.Recorder
}

func NewInteractor(
	appRepo domainrepository.AppRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	messageRepo domainrepository.MessageRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	poster MessagePoster,
	txManager domaintransaction.Manager,
	recorder audit.Recorder,
) *Interactor {
	return &Interactor{
		appRepo:           appRepo,
		userRepo:          userRepo,
		workspaceRepo:     workspaceRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		messageRepo:       messageRepo,
		channelAccessSvc:  channelAccessSvc,
		poster:            poster,
		txManager:         txManager,
		recorder:          recorder,
	}
}

// List はワークスペースのアプリを返します。公式アプリがまだなければ用意します
func (i *Interactor) List(ctx context.Context, input ListInput) ([]Output, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}
	if _, err := i.EnsureOfficial(ctx, input.WorkspaceID); err != nil {
		return nil, err
	}
	apps, err := i.appRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load apps: %w", err)
	}
	return i.toOutputs(ctx, apps, input.UserID, member.IsAdmin())
}

// ListByChannel はチャンネルに参加しているアプリを返します
func (i *Interactor) ListByChannel(ctx context.Context, input ListByChannelInput) ([]Output, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	apps, err := i.appRepo.FindByChannelID(ctx, ch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load apps: %w", err)
	}
	isAdmin, err := i.isAdmin(ctx, ch.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}
	return i.toOutputs(ctx, apps, input.UserID, isAdmin)
}

func (i *Interactor) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
	}
	app := &entity.App{WorkspaceID: input.WorkspaceID, CreatedBy: input.UserID}
	if err := i.applySettings(ctx, app, input.Settings, input.UserID); err != nil {
		return nil, err
	}
	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	app.TokenHash = &hash

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		bot, err := i.createBotUser(txCtx, app, false)
		if err != nil {
			return err
		}
		app.BotUserID = bot.ID
		if err := i.appRepo.Create(txCtx, app); err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}
		return i.joinDefaultChannel(txCtx, app)
	})
	if err != nil {
		return nil, err
	}
	i.record(ctx, app, input.UserID, entity.AuditActionAppCreated)

	outputs, err := i.toOutputs(ctx, []*entity.App{app}, input.UserID, false)
	if err != nil {
		return nil, err
	}
	return &CreateOutput{App: outputs[0], Token: token}, nil
}

func (i *Interactor) Update(ctx context.Context, input UpdateInput) (*Output, error) {
	app, err := i.findManageable(ctx, input.AppID, input.UserID)
	if err != nil {
		return nil, err
	}
	if err := i.applySettings(ctx, app, input.Settings, input.UserID); err != nil {
		return nil, err
	}

	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.appRepo.Update(txCtx, app); err != nil {
			return fmt.Errorf("failed to update app: %w", err)
		}
		// 過去の投稿の表示名も新しい名前にそろえる
		bot, err := i.userRepo.FindByID(txCtx, app.BotUserID)
		if err != nil {
			return fmt.Errorf("failed to load bot user: %w", err)
		}
		if bot != nil {
			bot.DisplayName = app.Name
			bot.AvatarURL = app.AvatarURL
			if err := i.userRepo.Update(txCtx, bot); err != nil {
				return fmt.Errorf("failed to update bot user: %w", err)
			}
		}
		return i.joinDefaultChannel(txCtx, app)
	})
	if err != nil {
		return nil, err
	}

	outputs, err := i.toOutputs(ctx, []*entity.App{app}, input.UserID, true)
	if err != nil {
		return nil, err
	}
	return &outputs[0], nil
}

func (i *Interactor) RegenerateToken(ctx context.Context, input TargetInput) (string, error) {
	app, err := i.findManageable(ctx, input.AppID, input.UserID)
	if err != nil {
		return "", err
	}
	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	app.TokenHash = &hash
	if err := i.appRepo.Update(ctx, app); err != nil {
		return "", fmt.Errorf("failed to update app: %w", err)
	}
	return token, nil
}

func (i *Interactor) Delete(ctx context.Context, input TargetInput) error {
	app, err := i.findManageable(ctx, input.AppID, input.UserID)
	if err != nil {
		return err
	}
	if err := i.appRepo.Delete(ctx, app.ID); err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	i.record(ctx, app, input.UserID, entity.AuditActionAppDeleted)
	return nil
}

// AddToChannel はアプリをチャンネルに参加させます。チャンネルの参加者でアプリを管理できる人だけができる
func (i *Interactor) AddToChannel(ctx context.Context, input ChannelInput) error {
	app, ch, err := i.findManageableInChannel(ctx, input)
	if err != nil {
		return err
	}
	if ch.ArchivedAt != nil {
		return domerr.ErrChannelArchived
	}
	return i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: ch.ID, UserID: app.BotUserID, Role: entity.ChannelRoleMember})
}

func (i *Interactor) RemoveFromChannel(ctx context.Context, input ChannelInput) error {
	app, ch, err := i.findManageableInChannel(ctx, input)
	if err != nil {
		return err
	}
	return i.channelMemberRepo.RemoveMember(ctx, ch.ID, app.BotUserID)
}

// Post は着信 Webhook に届いた内容をアプリのボットユーザー名義で投稿します
func (i *Interactor) Post(ctx context.Context, input PostInput) (*messageuc.MessageOutput, error) {
	if uuid.Validate(input.AppID) != nil {
		return nil, ErrAppNotFound
	}
	app, err := i.appRepo.FindByID(ctx, input.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to load app: %w", err)
	}
	// トークンの誤りも存在しない場合と区別しない
	if app == nil || !app.VerifyToken(input.Token) {
		return nil, ErrAppNotFound
	}
	creator, err := i.workspaceRepo.FindMember(ctx, app.WorkspaceID, app.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to verify creator: %w", err)
	}
	if creator == nil {
		return nil, ErrInactive
	}

	message, err := buildMessage(app, input)
	if err != nil {
		return nil, err
	}
	channelID := input.ChannelID
	if channelID == nil {
		channelID = app.DefaultChannelID
	}
	if channelID == nil {
		return nil, ErrChannelRequired
	}
	output, err := i.postAs(ctx, app, *channelID, input.ParentID, message)
	if err != nil {
		return nil, err
	}
	if err := i.appRepo.MarkUsed(ctx, app.ID, output.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to update last used time: %w", err)
	}
	return output, nil
}

// PostAsOfficial は公式アプリの名義で投稿します。投稿先を閲覧できるかは呼び出し側で確認済みであることが前提です
func (i *Interactor) PostAsOfficial(ctx context.Context, workspaceID, channelID string, parentID *string, body string) (*messageuc.MessageOutput, error) {
	official, err := i.EnsureOfficial(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return i.postAs(ctx, official, channelID, parentID, &entity.Message{UserID: official.BotUserID, Body: body})
}

// EnsureOfficial はワークスペースの公式アプリを返し、なければ作ります
func (i *Interactor) EnsureOfficial(ctx context.Context, workspaceID string) (*entity.App, error) {
	official, err := i.appRepo.FindOfficial(ctx, workspaceID)
	if err != nil || official != nil {
		return official, err
	}
	workspace, err := i.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, domerr.ErrNotFound
	}
	official = &entity.App{
		WorkspaceID: workspaceID,
		Name:        OfficialAppName,
		IsOfficial:  true,
		Permissions: []entity.AppPermission{entity.AppPermissionPostJoinedChannels, entity.AppPermissionPostPublicChannels, entity.AppPermissionPostThreadReplies},
		CreatedBy:   workspace.CreatedBy,
	}
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		bot, err := i.createBotUser(txCtx, official, true)
		if err != nil {
			return err
		}
		official.BotUserID = bot.ID
		return i.appRepo.Create(txCtx, official)
	})
	if err != nil {
		// 同時に作られたときは一意制約で失敗するため、先にできた方を使う
		if existing, findErr := i.appRepo.FindOfficial(ctx, workspaceID); findErr == nil && existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create official app: %w", err)
	}
	return official, nil
}

func (i *Interactor) postAs(ctx context.Context, app *entity.App, channelID string, parentID *string, message *entity.Message) (*messageuc.MessageOutput, error) {
	if uuid.Validate(channelID) != nil {
		return nil, domerr.ErrChannelNotFound
	}
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil || ch.WorkspaceID != app.WorkspaceID {
		return nil, domerr.ErrChannelNotFound
	}
	if ch.ArchivedAt != nil {
		return nil, domerr.ErrChannelArchived
	}
	isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, app.BotUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify channel membership: %w", err)
	}
	if !app.CanPostTo(ch, isMember) {
		return nil, ErrForbiddenChannel
	}
	if parentID != nil {
		if !app.IsOfficial && !app.Has(entity.AppPermissionPostThreadReplies) {
			return nil, ErrForbiddenThread
		}
		parent, err := i.messageRepo.FindByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != ch.ID || parent.ParentID != nil || parent.DeletedAt != nil {
			return nil, domerr.ErrParentMessageNotFound
		}
		message.ParentID = parentID
	}
	return i.poster.CreateBotMessage(ctx, ch, message)
}

// applySettings は入力を検証してアプリに反映します。既定のチャンネルは設定する人が参加しているものに限る
func (i *Interactor) applySettings(ctx context.Context, app *entity.App, settings SettingsInput, userID string) error {
	for _, p := range settings.Permissions {
		if !slices.Contains(entity.AllAppPermissions, p) {
			return ErrUnknownPermission
		}
	}
	if settings.AvatarURL != nil && !isHTTPURL(*settings.AvatarURL) {
		return ErrInvalidURL
	}
	if settings.OutgoingURL != nil && !isHTTPURL(*settings.OutgoingURL) {
		return ErrInvalidURL
	}
	if slices.Contains(settings.Permissions, entity.AppPermissionOutgoingWebhook) && settings.OutgoingURL == nil {
		return ErrOutgoingURLMissing
	}
	if id := settings.DefaultChannelID; id != nil && (app.DefaultChannelID == nil || *app.DefaultChannelID != *id) {
		ch, err := i.channelAccessSvc.EnsureChannelMember(ctx, *id, userID)
		if err != nil {
			return err
		}
		if ch.WorkspaceID != app.WorkspaceID {
			return domerr.ErrChannelNotFound
		}
		if ch.Type == entity.ChannelTypeDM || ch.Type == entity.ChannelTypeGroupDM {
			return ErrUnsupportedChannel
		}
	}

	app.Name = strings.TrimSpace(settings.Name)
	app.Description = settings.Description
	app.AvatarURL = settings.AvatarURL
	app.Permissions = slices.Compact(slices.Sorted(slices.Values(settings.Permissions)))
	app.DefaultChannelID = settings.DefaultChannelID
	app.OutgoingURL = settings.OutgoingURL
	switch {
	case settings.OutgoingURL == nil:
		app.OutgoingSecret = nil
	case app.OutgoingSecret == nil:
		secret, _, err := entity.NewSecretToken()
		if err != nil {
			return fmt.Errorf("failed to generate secret: %w", err)
		}
		app.OutgoingSecret = &secret
	}
	return nil
}

// createBotUser はアプリの投稿名義のユーザーを作ります。パスワードは照合できない値にし、メールアドレスは配送されない .invalid ドメインにする
func (i *Interactor) createBotUser(ctx context.Context, app *entity.App, isOfficial bool) (*entity.User, error) {
	bot := &entity.User{
		Email:        fmt.Sprintf("app-%s@app.invalid", uuid.NewString()),
		PasswordHash: entity.UnusablePasswordHash,
		DisplayName:  app.Name,
		AvatarURL:    app.AvatarURL,
		IsApp:        true,
		IsOfficial:   isOfficial,
	}
	if err := i.userRepo.Create(ctx, bot); err != nil {
		return nil, fmt.Errorf("failed to create bot user: %w", err)
	}
	return bot, nil
}

// joinDefaultChannel は既定のチャンネルにボットユーザーを参加させ、参加中のチャンネルとして投稿できるようにします
func (i *Interactor) joinDefaultChannel(ctx context.Context, app *entity.App) error {
	if app.DefaultChannelID == nil {
		return nil
	}
	return i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: *app.DefaultChannelID, UserID: app.BotUserID, Role: entity.ChannelRoleMember})
}

func buildMessage(app *entity.App, input PostInput) (*entity.Message, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return nil, ErrEmptyText
	}
	if utf8.RuneCountInString(text) > MaxTextLength {
		return nil, ErrTextTooLong
	}

	return &entity.Message{UserID: app.BotUserID, Body: text}, nil
}

// findManageable はアプリを取得し、作成者か管理者であることを確認します。公式アプリは誰も管理できない
func (i *Interactor) findManageable(ctx context.Context, appID, userID string) (*entity.App, error) {
	app, err := i.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to load app: %w", err)
	}
	if app == nil {
		return nil, ErrAppNotFound
	}
	if app.IsOfficial {
		return nil, ErrOfficialApp
	}
	if app.CreatedBy == userID {
		return app, nil
	}
	isAdmin, err := i.isAdmin(ctx, app.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, domerr.ErrUnauthorized
	}
	return app, nil
}

func (i *Interactor) findManageableInChannel(ctx context.Context, input ChannelInput) (*entity.App, *entity.Channel, error) {
	app, err := i.findManageable(ctx, input.AppID, input.UserID)
	if err != nil {
		return nil, nil, err
	}
	ch, err := i.channelAccessSvc.EnsureChannelMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, nil, err
	}
	if ch.WorkspaceID != app.WorkspaceID {
		return nil, nil, domerr.ErrChannelNotFound
	}
	if ch.Type == entity.ChannelTypeDM || ch.Type == entity.ChannelTypeGroupDM {
		return nil, nil, ErrUnsupportedChannel
	}
	return app, ch, nil
}

func (i *Interactor) isAdmin(ctx context.Context, workspaceID, userID string) (bool, error) {
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	return member != nil && member.IsAdmin(), nil
}

func (i *Interactor) toOutputs(ctx context.Context, apps []*entity.App, viewerID string, isAdmin bool) ([]Output, error) {
	creatorIDs := make([]string, 0, len(apps))
	for _, a := range apps {
		creatorIDs = append(creatorIDs, a.CreatedBy)
	}
	creators, err := i.userRepo.FindByIDs(ctx, creatorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load creators: %w", err)
	}
	byID := make(map[string]*entity.User, len(creators))
	for _, u := range creators {
		byID[u.ID] = u
	}

	outputs := make([]Output, 0, len(apps))
	for _, a := range apps {
		creator := messageuc.UserInfo{ID: a.CreatedBy}
		if u := byID[a.CreatedBy]; u != nil {
			creator = messageuc.UserInfo{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL}
		}
		canManage := !a.IsOfficial && (isAdmin || a.CreatedBy == viewerID)
		output := Output{
			ID:               a.ID,
			WorkspaceID:      a.WorkspaceID,
			Name:             a.Name,
			Description:      a.Description,
			AvatarURL:        a.AvatarURL,
			Permissions:      a.Permissions,
			DefaultChannelID: a.DefaultChannelID,
			OutgoingURL:      a.OutgoingURL,
			IsOfficial:       a.IsOfficial,
			BotUserID:        a.BotUserID,
			CreatedBy:        creator,
			CreatedAt:        a.CreatedAt,
			LastUsedAt:       a.LastUsedAt,
			CanManage:        canManage,
		}
		if canManage {
			output.OutgoingSecret = a.OutgoingSecret
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (i *Interactor) record(ctx context.Context, app *entity.App, actorID string, action entity.AuditAction) {
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: app.WorkspaceID,
		ActorID:     &actorID,
		Action:      action,
		TargetType:  entity.AuditTargetApp,
		TargetID:    app.ID,
		TargetLabel: app.Name,
	})
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
