package registry

import (
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	"gorm.io/gorm"
)

// Registry は分割されたRegistryを統合するメインのRegistryです
type Registry struct {
	domainRegistry         *DomainRegistry
	infrastructureRegistry *InfrastructureRegistry
	usecaseRegistry        *UseCaseRegistry
	interfaceRegistry      *InterfaceRegistry
}

// NewRegistry は新しいRegistryを作成します
func NewRegistry(db *gorm.DB, cfg *config.Config) *Registry {
	// ドメイン層のRegistryを作成
	domainRegistry := NewDomainRegistry(db)

	// WebSocketハブを作成
	hub := websocket.NewHub()

	// インフラストラクチャ層のRegistryを作成
	infrastructureRegistry := NewInfrastructureRegistry(db, cfg, hub, domainRegistry)

	// ユースケース層のRegistryを作成
	usecaseRegistry := NewUseCaseRegistry(domainRegistry, infrastructureRegistry)

	// インターフェース層のRegistryを作成
	interfaceRegistry := NewInterfaceRegistry(usecaseRegistry, infrastructureRegistry, domainRegistry)

	return &Registry{
		domainRegistry:         domainRegistry,
		infrastructureRegistry: infrastructureRegistry,
		usecaseRegistry:        usecaseRegistry,
		interfaceRegistry:      interfaceRegistry,
	}
}

// 各層のRegistryへのアクセサー
func (r *Registry) Domain() *DomainRegistry {
	return r.domainRegistry
}

func (r *Registry) Infrastructure() *InfrastructureRegistry {
	return r.infrastructureRegistry
}

func (r *Registry) UseCase() *UseCaseRegistry {
	return r.usecaseRegistry
}

func (r *Registry) Interface() *InterfaceRegistry {
	return r.interfaceRegistry
}

// 便利メソッド - 既存のコードとの互換性のため
func (r *Registry) NewRouter() *echo.Echo {
	return r.interfaceRegistry.NewRouter()
}

func (r *Registry) NewWebSocketHub() *websocket.Hub {
	return r.interfaceRegistry.NewWebSocketHub()
}
