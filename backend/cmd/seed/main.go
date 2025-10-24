package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/chat/internal/adapter/gateway/persistence"
	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/example/chat/internal/infrastructure/config"
	infradb "github.com/example/chat/internal/infrastructure/db"
	authuc "github.com/example/chat/internal/usecase/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	db, err := infradb.InitDB(cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := persistence.NewUserRepository(db)
	workspaceRepo := persistence.NewWorkspaceRepository(db)
	channelRepo := persistence.NewChannelRepository(db)
	messageRepo := persistence.NewMessageRepository(db)

	// Initialize password service for hashing
	passwordService := auth.NewPasswordService()

	// Check if seed data already exists
	if err := checkExistingData(db); err != nil {
		log.Fatalf("failed to check existing data: %v", err)
	}

	// Create seed data
	if err := createSeedData(userRepo, workspaceRepo, channelRepo, messageRepo, passwordService); err != nil {
		log.Fatalf("failed to create seed data: %v", err)
	}

	fmt.Println("✅ Seed data created successfully!")
}

func checkExistingData(db *gorm.DB) error {
	var count int64
	if err := db.Model(&infradb.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("⚠️  Database already contains data. Skipping seed data creation.")
		fmt.Println("   To reset the database, run: go run cmd/reset/main.go")
		return nil
	}

	return nil
}

func createSeedData(
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	messageRepo domainrepository.MessageRepository,
	passwordService authuc.PasswordService,
) error {
	ctx := context.Background()
	// Create test users
	users := []*infradb.User{
		{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Email:        "alice@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Alice Johnson",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=alice"),
		},
		{
			ID:           uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "bob@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Bob Smith",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=bob"),
		},
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Email:        "charlie@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Charlie Brown",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=charlie"),
		},
		{
			ID:           uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			Email:        "diana@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Diana Prince",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=diana"),
		},
	}

	// Create users
	for _, user := range users {
		if err := userRepo.Create(ctx, domainUserFromDB(user)); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
		fmt.Printf("✅ Created user: %s (%s)\n", user.DisplayName, user.Email)
	}

	// Create test workspace
	workspace := &infradb.Workspace{
		ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:        "Test Workspace",
		Description: stringPtr("A sample workspace for testing the chat application"),
		IconURL:     stringPtr("https://api.dicebear.com/7.x/initials/svg?seed=TW"),
		CreatedBy:   users[0].ID,
	}

	if err := workspaceRepo.Create(ctx, domainWorkspaceFromDB(workspace)); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	fmt.Printf("✅ Created workspace: %s\n", workspace.Name)

	// Add all users to workspace
	for i, user := range users {
		role := "member"
		if i == 0 {
			role = "owner"
		}

		member := &infradb.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      user.ID,
			Role:        role,
		}

		if err := workspaceRepo.AddMember(ctx, domainWorkspaceMemberFromDB(member)); err != nil {
			return fmt.Errorf("failed to add member %s to workspace: %w", user.DisplayName, err)
		}
		fmt.Printf("✅ Added %s as %s to workspace\n", user.DisplayName, role)
	}

	// Create channels
	channels := []*infradb.Channel{
		{
			ID:          uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			WorkspaceID: workspace.ID,
			Name:        "general",
			Description: stringPtr("General discussion channel"),
			IsPrivate:   false,
			CreatedBy:   users[0].ID,
		},
		{
			ID:          uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			WorkspaceID: workspace.ID,
			Name:        "random",
			Description: stringPtr("Random thoughts and off-topic discussions"),
			IsPrivate:   false,
			CreatedBy:   users[1].ID,
		},
		{
			ID:          uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
			WorkspaceID: workspace.ID,
			Name:        "development",
			Description: stringPtr("Development discussions and code reviews"),
			IsPrivate:   false,
			CreatedBy:   users[0].ID,
		},
		{
			ID:          uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
			WorkspaceID: workspace.ID,
			Name:        "private-team",
			Description: stringPtr("Private channel for team discussions"),
			IsPrivate:   true,
			CreatedBy:   users[0].ID,
		},
	}

	for _, channel := range channels {
		if err := channelRepo.Create(ctx, domainChannelFromDB(channel)); err != nil {
			return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
		}
		fmt.Printf("✅ Created channel: %s\n", channel.Name)

		// Add all users to public channels, only Alice and Bob to private channel
		usersToAdd := users
		if channel.IsPrivate {
			usersToAdd = users[:2] // Only Alice and Bob
		}

		for _, user := range usersToAdd {
			member := &infradb.ChannelMember{
				ChannelID: channel.ID,
				UserID:    user.ID,
			}

			if err := channelRepo.AddMember(ctx, domainChannelMemberFromDB(member)); err != nil {
				return fmt.Errorf("failed to add member %s to channel %s: %w", user.DisplayName, channel.Name, err)
			}
		}
		fmt.Printf("✅ Added %d members to channel %s\n", len(usersToAdd), channel.Name)
	}

	// Create sample messages
	messages := []*infradb.Message{
		// General channel messages
		{
			ID:        uuid.MustParse("f1111111-1111-1111-1111-111111111111"),
			ChannelID: channels[0].ID, // general
			UserID:    users[0].ID,    // Alice
			Body:      "👋 Welcome to our test workspace! This is a sample chat application.",
		},
		{
			ID:        uuid.MustParse("f2222222-2222-2222-2222-222222222222"),
			ChannelID: channels[0].ID, // general
			UserID:    users[1].ID,    // Bob
			Body:      "Hello everyone! Great to be here! 🎉",
		},
		{
			ID:        uuid.MustParse("f3333333-3333-3333-3333-333333333333"),
			ChannelID: channels[0].ID, // general
			UserID:    users[2].ID,    // Charlie
			Body:      "Thanks for the invite! Looking forward to working with everyone.",
		},
		{
			ID:        uuid.MustParse("f4444444-4444-4444-4444-444444444444"),
			ChannelID: channels[0].ID, // general
			UserID:    users[3].ID,    // Diana
			Body:      "Excited to be part of this team! 💪",
		},

		// Random channel messages
		{
			ID:        uuid.MustParse("f5555555-5555-5555-5555-555555555555"),
			ChannelID: channels[1].ID, // random
			UserID:    users[1].ID,    // Bob
			Body:      "Anyone else watching the latest season of that show? 🤔",
		},
		{
			ID:        uuid.MustParse("f6666666-6666-6666-6666-666666666666"),
			ChannelID: channels[1].ID, // random
			UserID:    users[2].ID,    // Charlie
			Body:      "Yes! The plot twist in episode 3 was incredible! 😱",
		},

		// Development channel messages
		{
			ID:        uuid.MustParse("f7777777-7777-7777-7777-777777777777"),
			ChannelID: channels[2].ID, // development
			UserID:    users[0].ID,    // Alice
			Body:      "I've pushed the latest changes to the main branch. Please review when you have time.",
		},
		{
			ID:        uuid.MustParse("f8888888-8888-8888-8888-888888888888"),
			ChannelID: channels[2].ID, // development
			UserID:    users[1].ID,    // Bob
			Body:      "I'll take a look at the PR. The new authentication flow looks solid! 👍",
		},
		{
			ID:        uuid.MustParse("f9999999-9999-9999-9999-999999999999"),
			ChannelID: channels[2].ID, // development
			UserID:    users[3].ID,    // Diana
			Body:      "Found a small issue with the mobile responsive design. I'll create a ticket for it.",
		},

		// Private team channel messages
		{
			ID:        uuid.MustParse("faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			ChannelID: channels[3].ID, // private-team
			UserID:    users[0].ID,    // Alice
			Body:      "Let's discuss the Q1 roadmap in this private channel.",
		},
		{
			ID:        uuid.MustParse("fbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			ChannelID: channels[3].ID, // private-team
			UserID:    users[1].ID,    // Bob
			Body:      "Sounds good! I think we should prioritize the user management features first.",
		},
	}

	// Create messages with staggered timestamps
	baseTime := time.Now().Add(-24 * time.Hour) // Start 24 hours ago
	for i, message := range messages {
		message.CreatedAt = baseTime.Add(time.Duration(i) * 30 * time.Minute)

		if err := messageRepo.Create(ctx, domainMessageFromDB(message)); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}

	fmt.Printf("✅ Created %d sample messages across all channels\n", len(messages))

	// Create some message reactions
	reactions := []*infradb.MessageReaction{
		{
			MessageID: messages[1].ID, // Bob's welcome message
			UserID:    users[0].ID,    // Alice
			Emoji:     "👋",
		},
		{
			MessageID: messages[1].ID, // Bob's welcome message
			UserID:    users[2].ID,    // Charlie
			Emoji:     "🎉",
		},
		{
			MessageID: messages[6].ID, // Bob's development message
			UserID:    users[0].ID,    // Alice
			Emoji:     "👍",
		},
	}

	for _, reaction := range reactions {
		if err := messageRepo.AddReaction(ctx, domainMessageReactionFromDB(reaction)); err != nil {
			return fmt.Errorf("failed to create message reaction: %w", err)
		}
	}

	fmt.Printf("✅ Created %d message reactions\n", len(reactions))

	return nil
}

// Helper functions for password hashing
func mustHashPassword(service authuc.PasswordService, password string) string {
	hash, err := service.HashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("failed to hash password: %v", err))
	}
	return hash
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// Domain conversion functions
func domainUserFromDB(user *infradb.User) *entity.User {
	return &entity.User{
		ID:           user.ID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func domainWorkspaceFromDB(workspace *infradb.Workspace) *entity.Workspace {
	return &entity.Workspace{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: workspace.Description,
		IconURL:     workspace.IconURL,
		CreatedBy:   workspace.CreatedBy.String(),
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

func domainWorkspaceMemberFromDB(member *infradb.WorkspaceMember) *entity.WorkspaceMember {
	return &entity.WorkspaceMember{
		WorkspaceID: member.WorkspaceID.String(),
		UserID:      member.UserID.String(),
		Role:        entity.WorkspaceRole(member.Role),
		JoinedAt:    member.JoinedAt,
	}
}

func domainChannelFromDB(channel *infradb.Channel) *entity.Channel {
	return &entity.Channel{
		ID:          channel.ID.String(),
		WorkspaceID: channel.WorkspaceID.String(),
		Name:        channel.Name,
		Description: channel.Description,
		IsPrivate:   channel.IsPrivate,
		CreatedBy:   channel.CreatedBy.String(),
		CreatedAt:   channel.CreatedAt,
		UpdatedAt:   channel.UpdatedAt,
	}
}

func domainChannelMemberFromDB(member *infradb.ChannelMember) *entity.ChannelMember {
	return &entity.ChannelMember{
		ChannelID: member.ChannelID.String(),
		UserID:    member.UserID.String(),
		JoinedAt:  member.JoinedAt,
	}
}

func domainMessageFromDB(message *infradb.Message) *entity.Message {
	var parentID *string
	if message.ParentID != nil {
		parentIDStr := message.ParentID.String()
		parentID = &parentIDStr
	}

	return &entity.Message{
		ID:        message.ID.String(),
		ChannelID: message.ChannelID.String(),
		UserID:    message.UserID.String(),
		ParentID:  parentID,
		Body:      message.Body,
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		DeletedAt: message.DeletedAt,
	}
}

func domainMessageReactionFromDB(reaction *infradb.MessageReaction) *entity.MessageReaction {
	return &entity.MessageReaction{
		MessageID: reaction.MessageID.String(),
		UserID:    reaction.UserID.String(),
		Emoji:     reaction.Emoji,
		CreatedAt: reaction.CreatedAt,
	}
}
