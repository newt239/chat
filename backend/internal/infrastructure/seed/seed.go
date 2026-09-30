package seed

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/repository"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// AutoSeed checks if the database is empty and seeds it with initial data
func AutoSeed(client *ent.Client) error {
	ctx := context.Background()

	// Check if database is empty
	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		// Check if the error is due to missing tables
		if strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("database tables do not exist. Please run migration first: %w", err)
		}
		return fmt.Errorf("failed to check user count: %w", err)
	}

	if userCount > 0 {
		log.Println("Database already contains data, skipping auto-seed")
		return nil
	}

	log.Println("Database is empty, seeding with initial data...")

	// Initialize repositories
	userRepo := repository.NewUserRepository(client)
	workspaceRepo := repository.NewWorkspaceRepository(client)
	channelRepo := repository.NewChannelRepository(client)
	channelMemberRepo := repository.NewChannelMemberRepository(client)
	messageRepo := repository.NewMessageRepository(client)

	// Initialize password service
	passwordService := auth.NewPasswordService()

	// Create seed data
	if err := createSeedData(ctx, client, userRepo, workspaceRepo, channelRepo, channelMemberRepo, messageRepo, passwordService); err != nil {
		return fmt.Errorf("failed to create seed data: %w", err)
	}

	log.Println("✅ Auto-seed completed successfully!")
	return nil
}

func createSeedData(
	ctx context.Context,
	client *ent.Client,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	messageRepo domainrepository.MessageRepository,
	passwordService authuc.PasswordService,
) error {
	// Create test users
	users := []*entity.User{
		{
			ID:           "11111111-1111-1111-1111-111111111111",
			Email:        "alice@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Alice Johnson",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=alice"),
		},
		{
			ID:           "22222222-2222-2222-2222-222222222222",
			Email:        "bob@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Bob Smith",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=bob"),
		},
		{
			ID:           "33333333-3333-3333-3333-333333333333",
			Email:        "charlie@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Charlie Brown",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=charlie"),
		},
		{
			ID:           "44444444-4444-4444-4444-444444444444",
			Email:        "diana@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  "Diana Prince",
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=diana"),
		},
	}

	// Create users
	for _, user := range users {
		if err := userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
	}

	// Create test workspaces (slug IDs)
	workspaces := []struct {
		ID          string
		Name        string
		Description string
		IsPublic    bool
		CreatedBy   string
	}{
		{ID: "general", Name: "General", Description: "一般的なディスカッション用のワークスペース", IsPublic: true, CreatedBy: users[0].ID},         // alice
		{ID: "engineering", Name: "Engineering", Description: "エンジニアリングチーム用のワークスペース", IsPublic: false, CreatedBy: users[1].ID}, // bob
		{ID: "marketing", Name: "Marketing", Description: "マーケティングチーム用のワークスペース", IsPublic: true, CreatedBy: users[2].ID},       // charlie
	}

	for _, ws := range workspaces {
		w := &entity.Workspace{
			ID:          ws.ID, // slug
			Name:        ws.Name,
			Description: stringPtr(ws.Description),
			IconURL:     nil,
			IsPublic:    ws.IsPublic,
			CreatedBy:   ws.CreatedBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := workspaceRepo.Create(ctx, w); err != nil {
			return fmt.Errorf("failed to create workspace %s: %w", ws.Name, err)
		}

		// Add creator as owner
		if err := workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{
			WorkspaceID: w.ID,
			UserID:      ws.CreatedBy,
			Role:        entity.WorkspaceRoleOwner,
			JoinedAt:    time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to add owner to workspace %s: %w", ws.Name, err)
		}
	}

	// Add additional members
	// General (public): add all users
	for i, user := range users {
		if i == 0 { // alice already owner
			continue
		}
		if err := workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{
			WorkspaceID: "general",
			UserID:      user.ID,
			Role:        entity.WorkspaceRoleMember,
			JoinedAt:    time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to add member to general workspace: %w", err)
		}
	}

	// Engineering (private): alice as admin in addition to bob(owner)
	if err := workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{
		WorkspaceID: "engineering",
		UserID:      users[0].ID, // alice
		Role:        entity.WorkspaceRoleAdmin,
		JoinedAt:    time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to add alice to engineering workspace: %w", err)
	}

	// Marketing (public): add alice as member in addition to charlie(owner)
	if err := workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{
		WorkspaceID: "marketing",
		UserID:      users[0].ID, // alice
		Role:        entity.WorkspaceRoleMember,
		JoinedAt:    time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to add alice to marketing workspace: %w", err)
	}

	channelDefinitions := []struct {
		id          string
		name        string
		description *string
		isPrivate   bool
		createdBy   string
		parentID    *string
	}{
		{
			id:          "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			name:        "general",
			description: stringPtr("General discussion channel"),
			isPrivate:   false,
			createdBy:   users[0].ID,
		},
		{
			id:          "cccccccc-cccc-cccc-cccc-cccccccccccc",
			name:        "random",
			description: stringPtr("Random thoughts and off-topic discussions"),
			isPrivate:   false,
			createdBy:   users[1].ID,
		},
		{
			id:          "dddddddd-dddd-dddd-dddd-dddddddddddd",
			name:        "development",
			description: stringPtr("Development discussions and code reviews"),
			isPrivate:   false,
			createdBy:   users[0].ID,
		},
		{
			id:          "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
			name:        "private-team",
			description: stringPtr("Private channel for team discussions"),
			isPrivate:   true,
			createdBy:   users[0].ID,
		},
		// 階層チャンネルの例
		{
			id:          "d0000000-0000-4000-8000-000000000001",
			name:        "dev",
			description: stringPtr("開発チーム全体"),
			createdBy:   users[0].ID,
		},
		{
			id:          "d0000000-0000-4000-8000-000000000002",
			name:        "dev/frontend",
			description: stringPtr("フロントエンド開発"),
			createdBy:   users[0].ID,
			parentID:    stringPtr("d0000000-0000-4000-8000-000000000001"),
		},
		{
			id:          "d0000000-0000-4000-8000-000000000003",
			name:        "dev/backend",
			description: stringPtr("バックエンド開発"),
			createdBy:   users[1].ID,
			parentID:    stringPtr("d0000000-0000-4000-8000-000000000001"),
		},
	}

	// Use the "general" workspace for default channels
	var channels []*entity.Channel
	for _, def := range channelDefinitions {
		channel, err := entity.NewChannel(entity.ChannelParams{
			ID:          def.id,
			WorkspaceID: "general",
			Name:        def.name,
			Description: def.description,
			IsPrivate:   def.isPrivate,
			ParentID:    def.parentID,
			CreatedBy:   def.createdBy,
		})
		if err != nil {
			return fmt.Errorf("failed to build channel %s: %w", def.name, err)
		}
		channels = append(channels, channel)

		if err := channelRepo.Create(ctx, channel); err != nil {
			return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
		}

		// Add all users to public channels, only Alice and Bob to private channel
		usersToAdd := users
		if channel.IsPrivate {
			usersToAdd = users[:2] // Only Alice and Bob
		}

		for _, user := range usersToAdd {
			member := &entity.ChannelMember{
				ChannelID: channel.ID,
				UserID:    user.ID,
			}

			if err := channelMemberRepo.AddMember(ctx, member); err != nil {
				return fmt.Errorf("failed to add member %s to channel %s: %w", user.DisplayName, channel.Name, err)
			}
		}
	}

	// Create sample messages
	messages := []*entity.Message{
		// General channel messages
		{
			ID:        "f1111111-1111-1111-1111-111111111111",
			ChannelID: channels[0].ID, // general
			UserID:    users[0].ID,    // Alice
			Body:      "👋 Welcome to our test workspace! This is a sample chat application.",
		},
		{
			ID:        "f2222222-2222-2222-2222-222222222222",
			ChannelID: channels[0].ID, // general
			UserID:    users[1].ID,    // Bob
			Body:      "Hello everyone! Great to be here! 🎉",
		},
		{
			ID:        "f3333333-3333-3333-3333-333333333333",
			ChannelID: channels[0].ID, // general
			UserID:    users[2].ID,    // Charlie
			Body:      "Thanks for the invite! Looking forward to working with everyone.",
		},
		{
			ID:        "f4444444-4444-4444-4444-444444444444",
			ChannelID: channels[0].ID, // general
			UserID:    users[3].ID,    // Diana
			Body:      "Excited to be part of this team! 💪",
		},

		// Random channel messages
		{
			ID:        "f5555555-5555-5555-5555-555555555555",
			ChannelID: channels[1].ID, // random
			UserID:    users[1].ID,    // Bob
			Body:      "Anyone else watching the latest season of that show? 🤔",
		},
		{
			ID:        "f6666666-6666-6666-6666-666666666666",
			ChannelID: channels[1].ID, // random
			UserID:    users[2].ID,    // Charlie
			Body:      "Yes! The plot twist in episode 3 was incredible! 😱",
		},

		// Development channel messages
		{
			ID:        "f7777777-7777-7777-7777-777777777777",
			ChannelID: channels[2].ID, // development
			UserID:    users[0].ID,    // Alice
			Body:      "I've pushed the latest changes to the main branch. Please review when you have time.",
		},
		{
			ID:        "f8888888-8888-8888-8888-888888888888",
			ChannelID: channels[2].ID, // development
			UserID:    users[1].ID,    // Bob
			Body:      "I'll take a look at the PR. The new authentication flow looks solid! 👍",
		},
		{
			ID:        "f9999999-9999-9999-9999-999999999999",
			ChannelID: channels[2].ID, // development
			UserID:    users[3].ID,    // Diana
			Body:      "Found a small issue with the mobile responsive design. I'll create a ticket for it.",
		},

		// Private team channel messages
		{
			ID:        "faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			ChannelID: channels[3].ID, // private-team
			UserID:    users[0].ID,    // Alice
			Body:      "Let's discuss the Q1 roadmap in this private channel.",
		},
		{
			ID:        "fbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			ChannelID: channels[3].ID, // private-team
			UserID:    users[1].ID,    // Bob
			Body:      "Sounds good! I think we should prioritize the user management features first.",
		},

		// 階層チャンネルのメッセージ
		{
			ID:        "d1000000-0000-4000-8000-000000000001",
			ChannelID: channels[4].ID, // dev
			UserID:    users[0].ID,    // Alice
			Body:      "dev の下にフロントエンドとバックエンドのチャンネルを作りました。",
		},
		{
			ID:        "d1000000-0000-4000-8000-000000000002",
			ChannelID: channels[5].ID, // dev/frontend
			UserID:    users[3].ID,    // Diana
			Body:      "サイドバーのツリー表示を実装中です。",
		},
		{
			ID:        "d1000000-0000-4000-8000-000000000003",
			ChannelID: channels[6].ID, // dev/backend
			UserID:    users[1].ID,    // Bob
			Body:      "親チャンネルで子孫のメッセージをまとめて取得できるようにしました。",
		},
	}

	// Create messages with staggered timestamps
	baseTime := time.Now().Add(-24 * time.Hour) // Start 24 hours ago
	for i, message := range messages {
		message.CreatedAt = baseTime.Add(time.Duration(i) * 30 * time.Minute)

		if err := messageRepo.Create(ctx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}

	// Create some message reactions
	reactions := []*entity.MessageReaction{
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
		if err := messageRepo.AddReaction(ctx, reaction); err != nil {
			return fmt.Errorf("failed to create message reaction: %w", err)
		}
	}

	// Create user groups
	userGroupRepo := repository.NewUserGroupRepository(client)
	groups := []*entity.UserGroup{
		{
			ID:          "0aaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			WorkspaceID: "general",
			Name:        "developers",
			Description: stringPtr("Development team members"),
			CreatedBy:   users[0].ID,
		},
		{
			ID:          "0bbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
			WorkspaceID: "general",
			Name:        "marketing",
			Description: stringPtr("Marketing team members"),
			CreatedBy:   users[0].ID,
		},
		{
			ID:          "0ccccccc-cccc-4ccc-8ccc-cccccccccccc",
			WorkspaceID: "general",
			Name:        "designers",
			Description: stringPtr("Design team members"),
			CreatedBy:   users[1].ID,
		},
	}

	for _, group := range groups {
		if err := userGroupRepo.Create(ctx, group); err != nil {
			return fmt.Errorf("failed to create user group %s: %w", group.Name, err)
		}
	}

	// Add members to groups
	groupMembers := []*entity.UserGroupMember{
		// developers group: Alice, Bob, Diana
		{GroupID: groups[0].ID, UserID: users[0].ID, JoinedAt: time.Now()},
		{GroupID: groups[0].ID, UserID: users[1].ID, JoinedAt: time.Now()},
		{GroupID: groups[0].ID, UserID: users[3].ID, JoinedAt: time.Now()},
		// marketing group: Bob, Charlie
		{GroupID: groups[1].ID, UserID: users[1].ID, JoinedAt: time.Now()},
		{GroupID: groups[1].ID, UserID: users[2].ID, JoinedAt: time.Now()},
		// designers group: Diana
		{GroupID: groups[2].ID, UserID: users[3].ID, JoinedAt: time.Now()},
	}

	for _, member := range groupMembers {
		if err := userGroupRepo.AddMember(ctx, member); err != nil {
			return fmt.Errorf("failed to add member to group: %w", err)
		}
	}

	// Create messages with mentions and links
	mentionMessages := []*entity.Message{
		{
			ID:        "fccccccc-cccc-cccc-cccc-cccccccccccc",
			ChannelID: channels[0].ID, // general
			UserID:    users[0].ID,    // Alice
			Body:      "Hey @bob, can you review the latest changes? Also check out this link: https://github.com/example/repo",
		},
		{
			ID:        "fddddddd-dddd-dddd-dddd-dddddddddddd",
			ChannelID: channels[0].ID, // general
			UserID:    users[1].ID,    // Bob
			Body:      "Sure @alice! @developers, let's discuss the new features. Here's a useful resource: https://docs.example.com/guide",
		},
		{
			ID:        "feeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
			ChannelID: channels[2].ID, // development
			UserID:    users[3].ID,    // Diana
			Body:      "@developers @designers, I've updated the UI mockups. Check this out: https://figma.com/design/example",
		},
	}

	// Create mention messages with timestamps
	for i, message := range mentionMessages {
		message.CreatedAt = baseTime.Add(time.Duration(len(messages)+i) * 30 * time.Minute)

		if err := messageRepo.Create(ctx, message); err != nil {
			return fmt.Errorf("failed to create mention message: %w", err)
		}
	}

	// Create user mentions using message repository
	userMentions := []*entity.MessageUserMention{
		{MessageID: mentionMessages[0].ID, UserID: users[1].ID, CreatedAt: mentionMessages[0].CreatedAt}, // Alice mentions Bob
		{MessageID: mentionMessages[1].ID, UserID: users[0].ID, CreatedAt: mentionMessages[1].CreatedAt}, // Bob mentions Alice
		{MessageID: mentionMessages[2].ID, UserID: users[0].ID, CreatedAt: mentionMessages[2].CreatedAt}, // Diana mentions Alice
		{MessageID: mentionMessages[2].ID, UserID: users[1].ID, CreatedAt: mentionMessages[2].CreatedAt}, // Diana mentions Bob
		{MessageID: mentionMessages[2].ID, UserID: users[3].ID, CreatedAt: mentionMessages[2].CreatedAt}, // Diana mentions Diana
	}

	for _, mention := range userMentions {
		if err := messageRepo.AddUserMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	// Create group mentions using message repository
	groupMentions := []*entity.MessageGroupMention{
		{MessageID: mentionMessages[1].ID, GroupID: groups[0].ID, CreatedAt: mentionMessages[1].CreatedAt}, // Bob mentions developers
		{MessageID: mentionMessages[2].ID, GroupID: groups[0].ID, CreatedAt: mentionMessages[2].CreatedAt}, // Diana mentions developers
		{MessageID: mentionMessages[2].ID, GroupID: groups[2].ID, CreatedAt: mentionMessages[2].CreatedAt}, // Diana mentions designers
	}

	for _, mention := range groupMentions {
		if err := messageRepo.AddGroupMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	// Create message links (simplified OGP data)
	linkRepo := repository.NewLinkRepository(client)
	links := []*entity.MessageLink{
		{
			MessageID: mentionMessages[0].ID,
			URL:       "https://github.com/example/repo",
			OGP: entity.OGPData{
				Title:       stringPtr("Example Repository"),
				Description: stringPtr("A sample repository for demonstration"),
				SiteName:    stringPtr("GitHub"),
			},
			CreatedAt: mentionMessages[0].CreatedAt,
		},
		{
			MessageID: mentionMessages[1].ID,
			URL:       "https://docs.example.com/guide",
			OGP: entity.OGPData{
				Title:       stringPtr("Developer Guide"),
				Description: stringPtr("Comprehensive guide for developers"),
				SiteName:    stringPtr("Example Docs"),
			},
			CreatedAt: mentionMessages[1].CreatedAt,
		},
		{
			MessageID: mentionMessages[2].ID,
			URL:       "https://figma.com/design/example",
			OGP: entity.OGPData{
				Title:       stringPtr("UI Design Mockups"),
				Description: stringPtr("Latest UI mockups for the project"),
				SiteName:    stringPtr("Figma"),
				CardType:    stringPtr("summary_large_image"),
			},
			CreatedAt: mentionMessages[2].CreatedAt,
		},
	}

	for _, link := range links {
		if err := linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create message link: %w", err)
		}
	}

	// Create a bookmark for Alice
	bookmarkRepo := repository.NewBookmarkRepository(client)
	bookmark := &entity.MessageBookmark{
		UserID:    users[0].ID,    // Alice
		MessageID: messages[1].ID, // Bob's welcome message
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	if err := bookmarkRepo.AddBookmark(ctx, bookmark); err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	if err := createDisplaySamples(ctx, client, messageRepo, users, channels, messages, baseTime.Add(time.Duration(len(messages)+len(mentionMessages))*30*time.Minute)); err != nil {
		return err
	}
	return createRichSamples(ctx, client, passwordService, users, channels, messages)
}

// createDisplaySamples はメッセージ表示のバリエーション（YouTube・メッセージリンク・ピン・多数のリアクション）を作ります。
// 添付ファイルは外部ストレージに実体が必要なため作りません
func createDisplaySamples(
	ctx context.Context,
	client *ent.Client,
	messageRepo domainrepository.MessageRepository,
	users []*entity.User,
	channels []*entity.Channel,
	messages []*entity.Message,
	startAt time.Time,
) error {
	permalink := func(msg *entity.Message) string {
		return fmt.Sprintf("http://localhost:5173/app/general/%s?message=%s", msg.ChannelID, msg.ID)
	}
	youtubeURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	durationSeconds := int32(213)
	thumbnailWidth, thumbnailHeight := int32(1280), int32(720)

	samples := []struct {
		message *entity.Message
		link    *entity.MessageLink
	}{
		{
			message: &entity.Message{ID: "f0c00001-0000-4000-8000-000000000001", ChannelID: channels[1].ID, UserID: users[2].ID, Body: "この動画がおすすめです " + youtubeURL},
			link: &entity.MessageLink{URL: youtubeURL, OGP: entity.OGPData{
				Title:       stringPtr("Rick Astley - Never Gonna Give You Up (Official Video)"),
				SiteName:    stringPtr("YouTube"),
				ImageURL:    stringPtr("https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg"),
				ImageWidth:  &thumbnailWidth,
				ImageHeight: &thumbnailHeight,
				YouTube:     &entity.YouTubeVideo{VideoID: "dQw4w9WgXcQ", ChannelName: stringPtr("Rick Astley"), DurationSeconds: &durationSeconds},
			}},
		},
		{
			message: &entity.Message{ID: "f0c00002-0000-4000-8000-000000000002", ChannelID: channels[0].ID, UserID: users[1].ID, Body: "レビュー依頼はこちらです " + permalink(messages[6])},
			link:    &entity.MessageLink{URL: permalink(messages[6]), LinkedMessageID: &messages[6].ID},
		},
		{
			// private-team のメッセージへのリンク。メンバーでない Charlie と Diana には引用カードが出ない
			message: &entity.Message{ID: "f0c00003-0000-4000-8000-000000000003", ChannelID: channels[0].ID, UserID: users[0].ID, Body: "ロードマップの議論はここを見てください " + permalink(messages[9])},
			link:    &entity.MessageLink{URL: permalink(messages[9]), LinkedMessageID: &messages[9].ID},
		},
	}

	linkRepo := repository.NewLinkRepository(client)
	for i, sample := range samples {
		sample.message.CreatedAt = startAt.Add(time.Duration(i) * 10 * time.Minute)
		if err := messageRepo.Create(ctx, sample.message); err != nil {
			return fmt.Errorf("failed to create sample message: %w", err)
		}
		sample.link.MessageID = sample.message.ID
		if err := linkRepo.Create(ctx, sample.link); err != nil {
			return fmt.Errorf("failed to create sample link: %w", err)
		}
	}

	pinRepo := repository.NewPinRepository(client)
	for _, pin := range []*entity.MessagePin{
		{ChannelID: messages[0].ChannelID, MessageID: messages[0].ID, PinnedBy: users[1].ID},
		{ChannelID: messages[6].ChannelID, MessageID: messages[6].ID, PinnedBy: users[0].ID},
	} {
		if err := pinRepo.Create(ctx, pin); err != nil {
			return fmt.Errorf("failed to create pin: %w", err)
		}
	}

	// ツールチップや「+N」の確認用に、1 つのメッセージへ多くのリアクションを付ける
	for i, emoji := range []string{"👍", "🎉", "❤️", "😂", "👀", "🚀", "✅", "🙏"} {
		for _, user := range users[:len(users)-i%len(users)] {
			reaction := &entity.MessageReaction{MessageID: messages[0].ID, UserID: user.ID, Emoji: emoji, CreatedAt: startAt.Add(time.Duration(i) * time.Minute)}
			if err := messageRepo.AddReaction(ctx, reaction); err != nil {
				return fmt.Errorf("failed to create sample reaction: %w", err)
			}
		}
	}

	return nil
}

// CreateSeedData creates seed data without checking if database is empty
func CreateSeedData(
	client *ent.Client,
	passwordService authuc.PasswordService,
) error {
	// Build repositories from client for simplicity
	userRepo := repository.NewUserRepository(client)
	workspaceRepo := repository.NewWorkspaceRepository(client)
	channelRepo := repository.NewChannelRepository(client)
	channelMemberRepo := repository.NewChannelMemberRepository(client)
	messageRepo := repository.NewMessageRepository(client)

	return createSeedData(context.Background(), client, userRepo, workspaceRepo, channelRepo, channelMemberRepo, messageRepo, passwordService)
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
