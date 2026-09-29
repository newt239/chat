package registry

import (
	"context"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/fcm"
	"github.com/newt239/chat/internal/infrastructure/link"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"github.com/newt239/chat/internal/infrastructure/mail"
	"github.com/newt239/chat/internal/infrastructure/meilisearch"
	"github.com/newt239/chat/internal/infrastructure/mention"
	"github.com/newt239/chat/internal/infrastructure/ogp"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/infrastructure/storage/local"
	"github.com/newt239/chat/internal/infrastructure/storage/wasabi"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	httphandler "github.com/newt239/chat/internal/interfaces/handler/http"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	invitationuc "github.com/newt239/chat/internal/usecase/invitation"
	notificationuc "github.com/newt239/chat/internal/usecase/notification"
)

// InfrastructureRegistry はインフラストラクチャ層の依存関係を管理します
type InfrastructureRegistry struct {
	client         *ent.Client
	config         *config.Config
	hub            *websocket.Hub
	redis          *goredis.Client
	ready          atomic.Bool
	domainRegistry *DomainRegistry
	messageIndex   *meilisearch.MessageIndex
	pushSender     notificationuc.Sender
}

// NewInfrastructureRegistry は新しいInfrastructureRegistryを作成します
func NewInfrastructureRegistry(client *ent.Client, cfg *config.Config, hub *websocket.Hub, rdb *goredis.Client, domainRegistry *DomainRegistry) *InfrastructureRegistry {
	r := &InfrastructureRegistry{
		client:         client,
		config:         cfg,
		hub:            hub,
		redis:          rdb,
		domainRegistry: domainRegistry,
		messageIndex:   meilisearch.NewMessageIndex(cfg.Search.MeilisearchURL, cfg.Search.MeilisearchAPIKey),
		pushSender:     newPushSender(cfg.Firebase.ProjectID),
	}
	r.ready.Store(true)
	return r
}

// Ready はリクエストを受け付けてよいかを返します。停止を始めたら false にする
func (r *InfrastructureRegistry) Ready() bool {
	return r.ready.Load()
}

func (r *InfrastructureRegistry) SetReady(ready bool) {
	r.ready.Store(ready)
}

// NewWebhookRateLimiter は Redis があれば全レプリカで共有して数えます。nil ならルーターがプロセス内で数える
func (r *InfrastructureRegistry) NewWebhookRateLimiter() httphandler.RateLimiter {
	if r.redis == nil {
		return nil
	}
	return redis.NewRateLimiter(r.redis, "webhook", httphandler.WebhookRatePerSecond, httphandler.WebhookBurst)
}

// newPushSender は FIREBASE_PROJECT_ID が未設定か初期化に失敗したら nil を返し、通知を送らない
func newPushSender(projectID string) notificationuc.Sender {
	if projectID == "" {
		return nil
	}
	sender, err := fcm.NewSender(context.Background(), projectID)
	if err != nil {
		logger.NewLogger().Warn("FCM を初期化できないためプッシュ通知を送りません", service.LogField{Key: "error", Value: err.Error()})
		return nil
	}
	return sender
}

// PushSender はプッシュ通知の送信役です。nil なら送らない
func (r *InfrastructureRegistry) PushSender() notificationuc.Sender {
	return r.pushSender
}

// Infrastructure Services
func (r *InfrastructureRegistry) NewJWTService() authuc.JWTService {
	return auth.NewJWTService(r.config.JWT.Secret)
}

func (r *InfrastructureRegistry) NewPasswordService() authuc.PasswordService {
	return auth.NewPasswordService()
}

func (r *InfrastructureRegistry) NewAuthSettings() authuc.Settings {
	return authuc.Settings{
		AccessTokenTTL:      time.Duration(r.config.JWT.AccessTokenTTL) * time.Minute,
		RefreshTokenTTL:     time.Duration(r.config.JWT.RefreshTokenTTL) * 24 * time.Hour,
		PasswordAuthEnabled: r.config.Auth.PasswordAuthEnabled,
	}
}

func (r *InfrastructureRegistry) NewGoogleVerifier() authuc.GoogleVerifier {
	return auth.NewGoogleVerifier(r.config.Auth.GoogleOAuthClientID)
}

func (r *InfrastructureRegistry) NewGoogleOAuth() *auth.GoogleOAuth {
	a := r.config.Auth
	return auth.NewGoogleOAuth(a.GoogleOAuthClientID, a.GoogleOAuthClientSecret, a.GoogleOAuthRedirectURL, a.NativeAppRedirectURL)
}

func (r *InfrastructureRegistry) NewInvitationSender() invitationuc.Sender {
	return mail.NoopInvitationSender{}
}

func (r *InfrastructureRegistry) NewNotificationService() *websocket.Notifier {
	return websocket.NewNotifier(r.hub)
}

func (r *InfrastructureRegistry) NewOGPService() service.OGPService {
	return ogp.NewOGPService()
}

func (r *InfrastructureRegistry) NewStorageService() service.StorageService {
	if r.config.Storage.Driver == "local" {
		return r.NewLocalStorage()
	}
	client, err := wasabi.NewClient(context.Background(), r.NewWasabiConfig())
	if err != nil {
		// エラーハンドリング: ログ出力してnilを返す
		// 実際のアプリケーションでは適切なエラーハンドリングが必要
		return nil
	}
	return wasabi.NewPresignService(client)
}

func (r *InfrastructureRegistry) NewStorageConfig() service.StorageConfig {
	if r.config.Storage.Driver == "local" {
		return r.NewLocalStorage()
	}
	return r.NewWasabiConfig()
}

// NewLocalStorage は開発用のストレージ。STORAGE_DRIVER=local のときだけ使う
func (r *InfrastructureRegistry) NewLocalStorage() *local.Storage {
	wasabiCfg := wasabi.NewConfig()
	return local.New(&local.Config{
		Dir:             r.config.Storage.LocalDir,
		BaseURL:         r.config.Storage.PublicBaseURL,
		Secret:          r.config.JWT.Secret,
		MaxFileSize:     wasabiCfg.MaxFileSize,
		UploadExpires:   wasabiCfg.UploadExpires,
		DownloadExpires: wasabiCfg.DownloadExpires,
	})
}

func (r *InfrastructureRegistry) NewWasabiConfig() *wasabi.Config {
	cfg := wasabi.NewConfig()
	cfg.Endpoint = r.config.Wasabi.Endpoint
	cfg.Region = r.config.Wasabi.Region
	cfg.AccessKeyID = r.config.Wasabi.AccessKeyID
	cfg.SecretAccessKey = r.config.Wasabi.SecretAccessKey
	cfg.BucketName = r.config.Wasabi.BucketName
	return cfg
}

func (r *InfrastructureRegistry) NewMentionService() service.MentionService {
	return mention.NewMentionService(
		r.domainRegistry.NewWorkspaceRepository(),
		r.domainRegistry.NewUserRepository(),
		r.domainRegistry.NewUserGroupRepository(),
		r.domainRegistry.NewMessageUserMentionRepository(),
		r.domainRegistry.NewMessageGroupMentionRepository(),
	)
}

func (r *InfrastructureRegistry) NewLinkProcessingService() service.LinkProcessingService {
	return link.NewLinkProcessingService(
		r.NewOGPService(),
		r.domainRegistry.NewMessageLinkRepository(),
		r.domainRegistry.NewMessageRepository(),
		r.domainRegistry.NewChannelRepository(),
	)
}

func (r *InfrastructureRegistry) NewTransactionManager() domaintransaction.Manager {
	return transaction.NewTransactionManager(r.client)
}

func (r *InfrastructureRegistry) MessageSearchIndex() *meilisearch.MessageIndex {
	return r.messageIndex
}

func (r *InfrastructureRegistry) NewLogger() service.Logger {
	return logger.NewLogger()
}
