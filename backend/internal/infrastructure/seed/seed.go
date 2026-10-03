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
	return CreateSeedData(ctx, client)
}

type seedMessage struct {
	id      string
	channel int
	user    int
	body    string
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
func CreateSeedData(ctx context.Context, client *ent.Client) error {
	passwordHash, err := auth.PasswordService{}.HashPassword("password123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
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
		user.PasswordHash = passwordHash
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
		member := &entity.UserGroupMember{GroupID: groups[m.group].ID, UserID: users[m.user].ID}
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
	}
	baseTime := time.Now().Add(-24 * time.Hour)
	messages := make([]*entity.Message, len(definitions))
	for i, def := range definitions {
		messages[i] = &entity.Message{ID: def.id, ChannelID: channels[def.channel].ID, CreatedAt: baseTime.Add(time.Duration(i) * 30 * time.Minute)}
		if err := client.Message.Create().
			SetID(uuid.MustParse(def.id)).
			SetChannelID(uuid.MustParse(messages[i].ChannelID)).
			SetUserID(uuid.MustParse(users[def.user].ID)).
			SetBody(def.body).
			SetCreatedAt(messages[i].CreatedAt).
			Exec(ctx); err != nil {
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

	if err := client.MessageBookmark.Create().
		SetUserID(uuid.MustParse(users[0].ID)).
		SetMessageID(uuid.MustParse(messages[1].ID)).
		SetCreatedAt(time.Now().Add(-2 * time.Hour)).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	return createRichSamples(ctx, client, passwordHash, users, channels, messages)
}

// samplePermalink はフロントの URL（APP_URL）でメッセージへのリンクを作ります
func samplePermalink(channelID, messageID string) string {
	return fmt.Sprintf("%s/app/general/%s?message=%s", os.Getenv("APP_URL"), channelID, messageID)
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
