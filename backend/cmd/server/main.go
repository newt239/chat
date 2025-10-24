package main

import (
	"log"
	"net/http"

	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/infrastructure/config"
	infradb "github.com/example/chat/internal/infrastructure/db"
	"github.com/example/chat/internal/infrastructure/logger"
	"github.com/example/chat/internal/infrastructure/repository"
	"github.com/example/chat/internal/infrastructure/seed"
	ginhttp "github.com/example/chat/internal/interface/http"
	"github.com/example/chat/internal/interface/http/handler"
	"github.com/example/chat/internal/interface/http/middleware"
	"github.com/example/chat/internal/interface/ws"
	authuc "github.com/example/chat/internal/usecase/auth"
	channeluc "github.com/example/chat/internal/usecase/channel"
	messageuc "github.com/example/chat/internal/usecase/message"
	reactionuc "github.com/example/chat/internal/usecase/reaction"
	readstateuc "github.com/example/chat/internal/usecase/readstate"
	workspaceuc "github.com/example/chat/internal/usecase/workspace"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: validate origin based on config
	},
}

func setupRouter(
	cfg *config.Config,
	jwtService *auth.JWTService,
	hub *ws.Hub,
	authHandler *handler.AuthHandler,
	workspaceHandler *handler.WorkspaceHandler,
	channelHandler *handler.ChannelHandler,
	messageHandler *handler.MessageHandler,
	readStateHandler *handler.ReadStateHandler,
	reactionHandler *handler.ReactionHandler,
) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		workspaceID := c.Query("workspaceId")
		if workspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspaceId required"})
			return
		}

		// Extract and validate JWT
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Sec-WebSocket-Protocol")
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		claims, err := jwtService.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Upgrade connection
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("websocket upgrade error: %v", err)
			return
		}

		wsConn := ws.NewConnection(hub, conn, claims.UserID, workspaceID)
		hub.Register(wsConn)

		go wsConn.WritePump()
		go wsConn.ReadPump()
	})

	// HTTP API routes
	ginhttp.RegisterRoutes(r, jwtService, authHandler, workspaceHandler, channelHandler, messageHandler, readStateHandler, reactionHandler)

	return r
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := infradb.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Auto-seed database if empty
	if err := seed.AutoSeed(db); err != nil {
		log.Fatalf("failed to auto-seed database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	readStateRepo := repository.NewReadStateRepository(db)

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT.Secret)
	passwordService := auth.NewPasswordService()

	// Initialize use cases
	authUseCase := authuc.NewAuthInteractor(userRepo, sessionRepo, jwtService, passwordService)
	workspaceUseCase := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)
	channelUseCase := channeluc.NewChannelInteractor(channelRepo, workspaceRepo)
	messageUseCase := messageuc.NewMessageInteractor(messageRepo, channelRepo, workspaceRepo, userRepo)
	readStateUseCase := readstateuc.NewReadStateInteractor(readStateRepo, channelRepo, workspaceRepo)
	reactionUseCase := reactionuc.NewReactionInteractor(messageRepo, channelRepo, workspaceRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceUseCase)
	channelHandler := handler.NewChannelHandler(channelUseCase)
	messageHandler := handler.NewMessageHandler(messageUseCase)
	readStateHandler := handler.NewReadStateHandler(readStateUseCase)
	reactionHandler := handler.NewReactionHandler(reactionUseCase)

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Setup and run server
	r := setupRouter(cfg, jwtService, hub, authHandler, workspaceHandler, channelHandler, messageHandler, readStateHandler, reactionHandler)
	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
