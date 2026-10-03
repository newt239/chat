package registry

import (
	"github.com/labstack/echo/v4"
	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

// Registry は各層の Registry をまとめ、起動に必要なものを組み立てます
type Registry struct {
	infrastructureRegistry *InfrastructureRegistry
	usecaseRegistry        *UseCaseRegistry
	interfaceRegistry      *InterfaceRegistry
}

// NewRegistry は rdb が nil なら WebSocket の配信などをプロセス内で完結させます
func NewRegistry(client *ent.Client, cfg *config.Config, rdb *goredis.Client) *Registry {
	domainRegistry := NewDomainRegistry(client)

	var hubOpts []websocket.HubOption
	if rdb != nil {
		hubOpts = append(hubOpts, websocket.WithBroker(redis.NewBroker(rdb)), websocket.WithPresenceStore(redis.NewPresenceStore(rdb)))
	}
	hub := websocket.NewHub(domainRegistry.NewChannelAccessService(), hubOpts...)

	infrastructureRegistry := NewInfrastructureRegistry(client, cfg, hub, rdb, domainRegistry)
	usecaseRegistry := NewUseCaseRegistry(domainRegistry, infrastructureRegistry)
	return &Registry{
		infrastructureRegistry: infrastructureRegistry,
		usecaseRegistry:        usecaseRegistry,
		interfaceRegistry:      NewInterfaceRegistry(usecaseRegistry, infrastructureRegistry, domainRegistry),
	}
}

func (r *Registry) Infrastructure() *InfrastructureRegistry {
	return r.infrastructureRegistry
}

func (r *Registry) UseCase() *UseCaseRegistry {
	return r.usecaseRegistry
}

func (r *Registry) NewRouter() *echo.Echo {
	return r.interfaceRegistry.NewRouter()
}

// Hub は全接続を持つ WebSocket のハブです
func (r *Registry) Hub() *websocket.Hub {
	return r.infrastructureRegistry.hub
}
