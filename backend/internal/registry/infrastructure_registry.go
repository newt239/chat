package registry

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/link"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"github.com/newt239/chat/internal/infrastructure/mention"
	"github.com/newt239/chat/internal/infrastructure/notification"
	"github.com/newt239/chat/internal/infrastructure/ogp"
	"github.com/newt239/chat/internal/infrastructure/storage/wasabi"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// InfrastructureRegistry はインフラストラクチャ層の依存関係を管理します
type InfrastructureRegistry struct {
	client         *ent.Client
	config         *config.Config
	hub            *websocket.Hub
	domainRegistry *DomainRegistry
}

// NewInfrastructureRegistry は新しいInfrastructureRegistryを作成します
func NewInfrastructureRegistry(client *ent.Client, cfg *config.Config, hub *websocket.Hub, domainRegistry *DomainRegistry) *InfrastructureRegistry {
	return &InfrastructureRegistry{
		client:         client,
		config:         cfg,
		hub:            hub,
		domainRegistry: domainRegistry,
	}
}

// Infrastructure Services
func (r *InfrastructureRegistry) NewJWTService() authuc.JWTService {
	return auth.NewJWTService(r.config.JWT.Secret)
}

func (r *InfrastructureRegistry) NewPasswordService() authuc.PasswordService {
	return auth.NewPasswordService()
}

func (r *InfrastructureRegistry) NewNotificationService() service.NotificationService {
	return notification.NewWebSocketNotificationService(r.hub)
}

func (r *InfrastructureRegistry) NewOGPService() service.OGPService {
	return ogp.NewOGPService()
}

func (r *InfrastructureRegistry) NewStorageService() service.StorageService {
	client, err := wasabi.NewClient(context.Background(), r.NewWasabiConfig())
	if err != nil {
		// エラーハンドリング: ログ出力してnilを返す
		// 実際のアプリケーションでは適切なエラーハンドリングが必要
		return nil
	}
	return wasabi.NewPresignService(client)
}

func (r *InfrastructureRegistry) NewStorageConfig() service.StorageConfig {
	return r.NewWasabiConfig()
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
	)
}

func (r *InfrastructureRegistry) NewTransactionManager() domaintransaction.Manager {
	return transaction.NewTransactionManager(r.client)
}

func (r *InfrastructureRegistry) NewLogger() service.Logger {
	return logger.NewLogger()
}
