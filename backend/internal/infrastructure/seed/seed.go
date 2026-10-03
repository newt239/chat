package seed

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/repository"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// developersGroupID はサンプルの本文からも参照する developers グループの ID
const developersGroupID = "0aaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

// AutoSeed はユーザーが 1 人もいないときだけ初期データを作ります
func AutoSeed(ctx context.Context, client *ent.Client) error {
	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check user count: %w", err)
	}
	if userCount > 0 {
		return nil
	}
	slog.Info("データベースが空のため初期データを作成します")
	return CreateSeedData(ctx, client, auth.PasswordService{})
}

type seedMessage struct {
	id      string
	channel int
	user    int
	body    string
}

func createMessage(ctx context.Context, client *ent.Client, m *entity.Message) error {
	return client.Message.Create().
		SetID(uuid.MustParse(m.ID)).
		SetChannelID(uuid.MustParse(m.ChannelID)).
		SetUserID(uuid.MustParse(m.UserID)).
		SetBody(m.Body).
		SetCreatedAt(m.CreatedAt).
		Exec(ctx)
}

func createReaction(ctx context.Context, client *ent.Client, messageID, userID, emoji string, at time.Time) error {
	return client.MessageReaction.Create().
		SetMessageID(uuid.MustParse(messageID)).
		SetUserID(uuid.MustParse(userID)).
		SetEmoji(emoji).
		SetCreatedAt(at).
		Exec(ctx)
}

// CreateSeedData は空のデータベースに確認用のユーザー・ワークスペース・チャンネル・メッセージを作ります
func CreateSeedData(ctx context.Context, client *ent.Client, passwordService authuc.PasswordService) error {
	userRepo := repository.NewUserRepository(client)
	workspaceRepo := repository.NewWorkspaceRepository(client)
	channelRepo := repository.NewChannelRepository(client)
	channelMemberRepo := repository.NewChannelMemberRepository(client)

	users := []*entity.User{
		{ID: "11111111-1111-1111-1111-111111111111", Email: "alice@example.com", DisplayName: "Alice Johnson"},
		{ID: "22222222-2222-2222-2222-222222222222", Email: "bob@example.com", DisplayName: "Bob Smith"},
		{ID: "33333333-3333-3333-3333-333333333333", Email: "charlie@example.com", DisplayName: "Charlie Brown"},
		{ID: "44444444-4444-4444-4444-444444444444", Email: "diana@example.com", DisplayName: "Diana Prince"},
	}
	for _, user := range users {
		user.PasswordHash = mustHashPassword(passwordService, "password123")
		user.AvatarURL = new("https://api.dicebear.com/7.x/avataaars/svg?seed=" + user.Email[:len(user.Email)-len("@example.com")])
		if err := userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
	}

	workspaces := []entity.Workspace{
		{ID: "general", Name: "General", Description: new("一般的なディスカッション用のワークスペース"), IsPublic: true, CreatedBy: users[0].ID},
		{ID: "engineering", Name: "Engineering", Description: new("エンジニアリングチーム用のワークスペース"), CreatedBy: users[1].ID},
		{ID: "marketing", Name: "Marketing", Description: new("マーケティングチーム用のワークスペース"), IsPublic: true, CreatedBy: users[2].ID},
	}
	members := []entity.WorkspaceMember{
		{WorkspaceID: "general", UserID: users[1].ID, Role: entity.WorkspaceRoleMember},
		{WorkspaceID: "general", UserID: users[2].ID, Role: entity.WorkspaceRoleMember},
		{WorkspaceID: "general", UserID: users[3].ID, Role: entity.WorkspaceRoleMember},
		{WorkspaceID: "engineering", UserID: users[0].ID, Role: entity.WorkspaceRoleAdmin},
		{WorkspaceID: "marketing", UserID: users[0].ID, Role: entity.WorkspaceRoleMember},
	}
	for _, ws := range workspaces {
		if err := workspaceRepo.Create(ctx, &ws); err != nil {
			return fmt.Errorf("failed to create workspace %s: %w", ws.Name, err)
		}
		members = append(members, entity.WorkspaceMember{WorkspaceID: ws.ID, UserID: ws.CreatedBy, Role: entity.WorkspaceRoleOwner})
	}
	for _, member := range members {
		member.JoinedAt = time.Now()
		if err := workspaceRepo.AddMember(ctx, &member); err != nil {
			return fmt.Errorf("failed to add member to workspace %s: %w", member.WorkspaceID, err)
		}
	}

	devID := "d0000000-0000-4000-8000-000000000001"
	channelDefinitions := []entity.Channel{
		{ID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "general", Description: new("General discussion channel"), CreatedBy: users[0].ID},
		{ID: "cccccccc-cccc-cccc-cccc-cccccccccccc", Name: "random", Description: new("Random thoughts and off-topic discussions"), CreatedBy: users[1].ID},
		{ID: "dddddddd-dddd-dddd-dddd-dddddddddddd", Name: "development", Description: new("Development discussions and code reviews"), CreatedBy: users[0].ID},
		{ID: "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", Name: "private-team", Description: new("Private channel for team discussions"), Type: entity.ChannelTypePrivate, CreatedBy: users[0].ID},
		{ID: devID, Name: "dev", Description: new("開発チーム全体"), CreatedBy: users[0].ID},
		{ID: "d0000000-0000-4000-8000-000000000002", Name: "dev/frontend", Description: new("フロントエンド開発"), ParentID: &devID, CreatedBy: users[0].ID},
		{ID: "d0000000-0000-4000-8000-000000000003", Name: "dev/backend", Description: new("バックエンド開発"), ParentID: &devID, CreatedBy: users[1].ID},
	}
	var channels []*entity.Channel
	for _, def := range channelDefinitions {
		def.WorkspaceID = "general"
		channel, err := entity.NewChannel(def)
		if err != nil {
			return fmt.Errorf("failed to build channel %s: %w", def.Name, err)
		}
		if err := channelRepo.Create(ctx, channel); err != nil {
			return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
		}
		channels = append(channels, channel)
		// 非公開チャンネルには Alice と Bob だけを入れる
		usersToAdd := users
		if channel.IsPrivate() {
			usersToAdd = users[:2]
		}
		for _, user := range usersToAdd {
			if err := channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: channel.ID, UserID: user.ID}); err != nil {
				return fmt.Errorf("failed to add member %s to channel %s: %w", user.DisplayName, channel.Name, err)
			}
		}
	}

	groups := []*entity.UserGroup{
		{ID: developersGroupID, WorkspaceID: "general", Name: "developers", Description: new("Development team members"), CreatedBy: users[0].ID},
		{ID: "0bbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", WorkspaceID: "general", Name: "marketing", Description: new("Marketing team members"), CreatedBy: users[0].ID},
		{ID: "0ccccccc-cccc-4ccc-8ccc-cccccccccccc", WorkspaceID: "general", Name: "designers", Description: new("Design team members"), CreatedBy: users[1].ID},
	}
	userGroupRepo := repository.NewUserGroupRepository(client)
	for _, group := range groups {
		if err := userGroupRepo.Create(ctx, group); err != nil {
			return fmt.Errorf("failed to create user group %s: %w", group.Name, err)
		}
	}
	for _, m := range []struct{ group, user int }{{0, 0}, {0, 1}, {0, 3}, {1, 1}, {1, 2}, {2, 3}} {
		member := &entity.UserGroupMember{GroupID: groups[m.group].ID, UserID: users[m.user].ID, JoinedAt: time.Now()}
		if err := userGroupRepo.AddMember(ctx, member); err != nil {
			return fmt.Errorf("failed to add member to group: %w", err)
		}
	}

	definitions := []seedMessage{
		{id: "f1111111-1111-1111-1111-111111111111", channel: 0, user: 0, body: "👋 Welcome to our test workspace! This is a sample chat application."},
		{id: "f2222222-2222-2222-2222-222222222222", channel: 0, user: 1, body: "Hello everyone! Great to be here! 🎉"},
		{id: "f3333333-3333-3333-3333-333333333333", channel: 0, user: 2, body: "Thanks for the invite! Looking forward to working with everyone."},
		{id: "f4444444-4444-4444-4444-444444444444", channel: 0, user: 3, body: "Excited to be part of this team! 💪"},
		{id: "f5555555-5555-5555-5555-555555555555", channel: 1, user: 1, body: "Anyone else watching the latest season of that show? 🤔"},
		{id: "f6666666-6666-6666-6666-666666666666", channel: 1, user: 2, body: "Yes! The plot twist in episode 3 was incredible! 😱"},
		{id: "f7777777-7777-7777-7777-777777777777", channel: 2, user: 0, body: "I've pushed the latest changes to the main branch. Please review when you have time."},
		{id: "f8888888-8888-8888-8888-888888888888", channel: 2, user: 1, body: "I'll take a look at the PR. The new authentication flow looks solid! 👍"},
		{id: "f9999999-9999-9999-9999-999999999999", channel: 2, user: 3, body: "Found a small issue with the mobile responsive design. I'll create a ticket for it."},
		{id: "faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", channel: 3, user: 0, body: "Let's discuss the Q1 roadmap in this private channel."},
		{id: "fbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", channel: 3, user: 1, body: "Sounds good! I think we should prioritize the user management features first."},
		{id: "d1000000-0000-4000-8000-000000000001", channel: 4, user: 0, body: "dev の下にフロントエンドとバックエンドのチャンネルを作りました。"},
		{id: "d1000000-0000-4000-8000-000000000002", channel: 5, user: 3, body: "サイドバーのツリー表示を実装中です。"},
		{id: "d1000000-0000-4000-8000-000000000003", channel: 6, user: 1, body: "親チャンネルで子孫のメッセージをまとめて取得できるようにしました。"},
		{id: "fccccccc-cccc-cccc-cccc-cccccccccccc", channel: 0, user: 0, body: "Hey <@" + users[1].ID + ">, can you review the latest changes? Also check out this link: https://github.com/example/repo"},
		{id: "fddddddd-dddd-dddd-dddd-dddddddddddd", channel: 0, user: 1, body: "Sure <@" + users[0].ID + ">! <@&" + groups[0].ID + ">, let's discuss the new features. Here's a useful resource: https://docs.example.com/guide"},
		{id: "feeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", channel: 2, user: 3, body: "<@&" + groups[0].ID + "> <@&" + groups[2].ID + ">, I've updated the UI mockups. Check this out: https://figma.com/design/example"},
	}
	baseTime := time.Now().Add(-24 * time.Hour)
	messages := make([]*entity.Message, len(definitions))
	for i, def := range definitions {
		messages[i] = &entity.Message{ID: def.id, ChannelID: channels[def.channel].ID, UserID: users[def.user].ID, Body: def.body, CreatedAt: baseTime.Add(time.Duration(i) * 30 * time.Minute)}
		if err := createMessage(ctx, client, messages[i]); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}

	for _, r := range []struct {
		message, user int
		emoji         string
	}{{1, 0, "👋"}, {1, 2, "🎉"}, {6, 0, "👍"}} {
		if err := createReaction(ctx, client, messages[r.message].ID, users[r.user].ID, r.emoji, messages[r.message].CreatedAt.Add(time.Minute)); err != nil {
			return fmt.Errorf("failed to create message reaction: %w", err)
		}
	}

	// グループへのメンションは投稿時点のメンバーに展開して保存する
	mentionMessages := messages[14:]
	developers := &groups[0].ID
	userMentions := []*entity.MessageUserMention{
		{MessageID: mentionMessages[0].ID, UserID: users[1].ID},
		{MessageID: mentionMessages[1].ID, UserID: users[0].ID},
		{MessageID: mentionMessages[1].ID, UserID: users[1].ID, ViaGroupID: developers},
		{MessageID: mentionMessages[1].ID, UserID: users[3].ID, ViaGroupID: developers},
		{MessageID: mentionMessages[2].ID, UserID: users[0].ID, ViaGroupID: developers},
		{MessageID: mentionMessages[2].ID, UserID: users[1].ID, ViaGroupID: developers},
		{MessageID: mentionMessages[2].ID, UserID: users[3].ID, ViaGroupID: developers},
	}
	if err := repository.NewMessageUserMentionRepository(client).CreateBulk(ctx, userMentions); err != nil {
		return fmt.Errorf("failed to create user mentions: %w", err)
	}
	groupMentions := []*entity.MessageGroupMention{
		{MessageID: mentionMessages[1].ID, GroupID: groups[0].ID},
		{MessageID: mentionMessages[2].ID, GroupID: groups[0].ID},
		{MessageID: mentionMessages[2].ID, GroupID: groups[2].ID},
	}
	if err := repository.NewMessageGroupMentionRepository(client).CreateBulk(ctx, groupMentions); err != nil {
		return fmt.Errorf("failed to create group mentions: %w", err)
	}

	linkRepo := repository.NewLinkRepository(client)
	links := []*entity.MessageLink{
		{MessageID: mentionMessages[0].ID, URL: "https://github.com/example/repo", OGP: entity.OGPData{
			Title: new("Example Repository"), Description: new("A sample repository for demonstration"), SiteName: new("GitHub"),
		}},
		{MessageID: mentionMessages[1].ID, URL: "https://docs.example.com/guide", OGP: entity.OGPData{
			Title: new("Developer Guide"), Description: new("Comprehensive guide for developers"), SiteName: new("Example Docs"),
		}},
		{MessageID: mentionMessages[2].ID, URL: "https://figma.com/design/example", OGP: entity.OGPData{
			Title: new("UI Design Mockups"), Description: new("Latest UI mockups for the project"), SiteName: new("Figma"), CardType: new("summary_large_image"),
		}},
	}
	for _, link := range links {
		if err := createLink(ctx, linkRepo, link); err != nil {
			return fmt.Errorf("failed to create message link: %w", err)
		}
	}

	if err := client.MessageBookmark.Create().
		SetUserID(uuid.MustParse(users[0].ID)).
		SetMessageID(uuid.MustParse(messages[1].ID)).
		SetCreatedAt(time.Now().Add(-2 * time.Hour)).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	if err := createDisplaySamples(ctx, client, users, channels, messages, baseTime.Add(time.Duration(len(messages))*30*time.Minute)); err != nil {
		return err
	}
	return createRichSamples(ctx, client, passwordService, users, channels, messages)
}

// createDisplaySamples は YouTube・メッセージリンク・ピン・多数のリアクションの表示確認用データを作ります
func createDisplaySamples(ctx context.Context, client *ent.Client, users []*entity.User, channels []*entity.Channel, messages []*entity.Message, startAt time.Time) error {
	permalink := func(msg *entity.Message) string {
		return samplePermalink(msg.ChannelID, msg.ID)
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
				Title:       new("Rick Astley - Never Gonna Give You Up (Official Video)"),
				SiteName:    new("YouTube"),
				ImageURL:    new("https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg"),
				ImageWidth:  &thumbnailWidth,
				ImageHeight: &thumbnailHeight,
				YouTube:     &entity.YouTubeVideo{VideoID: "dQw4w9WgXcQ", ChannelName: new("Rick Astley"), DurationSeconds: &durationSeconds},
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
		if err := createMessage(ctx, client, sample.message); err != nil {
			return fmt.Errorf("failed to create sample message: %w", err)
		}
		sample.link.MessageID = sample.message.ID
		if err := createLink(ctx, linkRepo, sample.link); err != nil {
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
			if err := createReaction(ctx, client, messages[0].ID, user.ID, emoji, startAt.Add(time.Duration(i)*time.Minute)); err != nil {
				return fmt.Errorf("failed to create sample reaction: %w", err)
			}
		}
	}
	return nil
}

// samplePermalink はフロントの URL（APP_URL）でメッセージへのリンクを作ります
func samplePermalink(channelID, messageID string) string {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://chat.localhost"
	}
	return fmt.Sprintf("%s/app/general/%s?message=%s", appURL, channelID, messageID)
}

func mustHashPassword(service authuc.PasswordService, password string) string {
	hash, err := service.HashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("failed to hash password: %v", err))
	}
	return hash
}

// createLink はメッセージへのリンクでなければプレビューを先に保存してからリンクを作ります
func createLink(ctx context.Context, linkRepo domainrepository.MessageLinkRepository, link *entity.MessageLink) error {
	if link.LinkedMessageID == nil {
		preview := &entity.LinkPreview{URL: link.URL, OGP: link.OGP, FetchedAt: time.Now()}
		if err := linkRepo.UpsertPreview(ctx, preview); err != nil {
			return err
		}
		link.LinkPreviewID = &preview.ID
	}
	return linkRepo.CreateBulk(ctx, []*entity.MessageLink{link})
}
