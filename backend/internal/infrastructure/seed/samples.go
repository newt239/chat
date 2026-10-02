package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/repository"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

const scrollTestMessages = 2000

// 確認用に作るチャンネル。alice が参加しないものは参加前のプレビューや「チャンネルに参加」の確認に使う
var sampleChannels = []struct {
	name        string
	description string
	parentName  string
	aliceJoins  bool
}{
	{name: "showcase", description: "いろいろな種類のメッセージの表示確認用", aliceJoins: true},
	{name: "scroll-test", description: "バーチャルスクロールの確認用（2,000 件）", aliceJoins: true},
	{name: "announcements", description: "全体へのお知らせ", aliceJoins: true},
	{name: "dev/frontend/web", description: "Web フロントエンド", parentName: "dev/frontend", aliceJoins: true},
	{name: "dev/frontend/mobile", description: "モバイルアプリ", parentName: "dev/frontend", aliceJoins: true},
	{name: "design", description: "デザインの相談（alice は未参加）"},
	{name: "help", description: "困ったときの質問窓口（alice は未参加）"},
}

var sampleUserNames = []string{
	"Eve Tanaka", "Frank Suzuki", "Grace Sato", "Hiro Yamada", "Ivy Kobayashi", "Jun Ito", "Ken Watanabe", "山田 花子",
}

// createRichSamples はメッセージの種類・システムメッセージ・スレッド・大量のメッセージなど、
// 画面の確認に使うデータを追加します。alice の視点で並び順・ミュート・スターも確認できるようにします
func createRichSamples(
	ctx context.Context,
	client *ent.Client,
	passwordService authuc.PasswordService,
	users []*entity.User,
	channels []*entity.Channel,
	messages []*entity.Message,
) error {
	alice := users[0]
	members, err := createSampleUsers(ctx, passwordService, client, users)
	if err != nil {
		return err
	}
	byName := make(map[string]*entity.Channel, len(channels))
	for _, ch := range channels {
		byName[ch.Name] = ch
	}
	// 既存の公開チャンネルにも新しいメンバーを入れ、参加のお知らせを残す
	joinedAt := time.Now().Add(-26 * time.Hour)
	for _, ch := range channels {
		if ch.IsPrivate {
			continue
		}
		for i, member := range members[len(users):] {
			if err := addSampleMember(ctx, client, ch, member, joinedAt.Add(time.Duration(i)*time.Minute)); err != nil {
				return err
			}
		}
	}
	for _, def := range sampleChannels {
		params := entity.ChannelParams{
			ID:          uuid.NewString(),
			WorkspaceID: "general",
			Name:        def.name,
			Description: stringPtr(def.description),
			CreatedBy:   users[1].ID,
		}
		if parent, ok := byName[def.parentName]; ok {
			params.ParentID = &parent.ID
		}
		ch, err := entity.NewChannel(params)
		if err != nil {
			return fmt.Errorf("failed to build channel %s: %w", def.name, err)
		}
		if err := repository.NewChannelRepository(client).Create(ctx, ch); err != nil {
			return fmt.Errorf("failed to create channel %s: %w", def.name, err)
		}
		byName[ch.Name] = ch
		for i, member := range members {
			if member.ID == alice.ID && !def.aliceJoins {
				continue
			}
			if err := addSampleMember(ctx, client, ch, member, joinedAt.Add(time.Duration(i)*time.Minute)); err != nil {
				return err
			}
		}
	}

	if err := createShowcaseMessages(ctx, client, members, byName, messages); err != nil {
		return err
	}
	if err := createScrollTestMessages(ctx, client, members, byName["scroll-test"]); err != nil {
		return err
	}
	for _, name := range []string{"design", "help", "announcements", "dev/frontend/web"} {
		if err := createChatter(ctx, client, members, byName[name]); err != nil {
			return err
		}
	}

	// alice の視点での並び順の確認用。random と announcements はミュート、showcase と development はスター
	for _, name := range []string{"random", "announcements"} {
		if err := client.ChannelMute.Create().SetUserID(uuid.MustParse(alice.ID)).SetChannelID(uuid.MustParse(byName[name].ID)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to mute channel: %w", err)
		}
	}
	for _, name := range []string{"showcase", "development"} {
		if err := client.ChannelStar.Create().SetUserID(uuid.MustParse(alice.ID)).SetChannelID(uuid.MustParse(byName[name].ID)).Exec(ctx); err != nil {
			return fmt.Errorf("failed to star channel: %w", err)
		}
	}
	return nil
}

// createSampleUsers は既存のユーザーに確認用のユーザーを足して返します。日本語だけの名前はメンションできない例です
func createSampleUsers(ctx context.Context, passwordService authuc.PasswordService, client *ent.Client, users []*entity.User) ([]*entity.User, error) {
	workspaceRepo := repository.NewWorkspaceRepository(client)
	userRepo := repository.NewUserRepository(client)
	members := append([]*entity.User{}, users...)
	for i, name := range sampleUserNames {
		seed := strings.ToLower(strings.Fields(name)[0])
		if i == len(sampleUserNames)-1 {
			seed = "hanako"
		}
		user := &entity.User{
			ID:           uuid.NewString(),
			Email:        seed + "@example.com",
			PasswordHash: mustHashPassword(passwordService, "password123"),
			DisplayName:  name,
			AvatarURL:    stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=" + seed),
		}
		if err := userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", name, err)
		}
		if err := workspaceRepo.AddMember(ctx, &entity.WorkspaceMember{
			WorkspaceID: "general",
			UserID:      user.ID,
			Role:        entity.WorkspaceRoleMember,
			JoinedAt:    time.Now(),
		}); err != nil {
			return nil, fmt.Errorf("failed to add %s to workspace: %w", name, err)
		}
		members = append(members, user)
	}
	return members, nil
}

// addSampleMember はチャンネルに参加させ、参加のお知らせを残します
func addSampleMember(ctx context.Context, client *ent.Client, ch *entity.Channel, user *entity.User, at time.Time) error {
	if err := repository.NewChannelMemberRepository(client).AddMember(ctx, &entity.ChannelMember{ChannelID: ch.ID, UserID: user.ID}); err != nil {
		return fmt.Errorf("failed to add %s to %s: %w", user.DisplayName, ch.Name, err)
	}
	return createSystemMessage(ctx, client, ch, entity.SystemMessageKindMemberJoined, user.ID, map[string]any{"actorId": user.ID, "userId": user.ID}, at)
}

func createSystemMessage(ctx context.Context, client *ent.Client, ch *entity.Channel, kind entity.SystemMessageKind, actorID string, payload map[string]any, at time.Time) error {
	if err := client.SystemMessage.Create().
		SetChannelID(uuid.MustParse(ch.ID)).
		SetActorID(uuid.MustParse(actorID)).
		SetKind(string(kind)).
		SetPayload(payload).
		SetCreatedAt(at).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to create system message: %w", err)
	}
	return nil
}

type sampleMessage struct {
	userIndex int
	body      string
	edited    bool
	deleted   bool
	location  *entity.MessageLocation
	// 付けるリアクションと、リアクションするユーザーの数
	reactions map[string]int
	replies   []string
	link      *entity.MessageLink
}

// createShowcaseMessages は Markdown・コード・メンション・リンク・位置情報・編集・削除・スレッド・ピンなどを 1 つのチャンネルに並べます
func createShowcaseMessages(ctx context.Context, client *ent.Client, users []*entity.User, channelsByName map[string]*entity.Channel, messages []*entity.Message) error {
	ch := channelsByName["showcase"]
	general := messages[0]
	permalink := samplePermalink(general.ChannelID, general.ID)
	longText := strings.Repeat("長いメッセージの折り返しと高さの確認用の文章です。仮想スクロールでは行ごとに高さが変わるため、長文が混ざっても位置がずれないことを確かめます。", 6)
	samples := []sampleMessage{
		{userIndex: 1, body: "# 見出し 1\n## 見出し 2\n\n**太字**・*斜体*・~~取り消し~~・`インラインコード`\n\n- 箇条書き\n  - 入れ子\n- [ ] タスク\n- [x] 完了したタスク\n\n1. 番号付き\n2. リスト"},
		{userIndex: 2, body: "> 引用です。\n> 複数行の引用も表示できます。\n\n| 項目 | 状態 | 担当 |\n| --- | --- | --- |\n| ログイン | 完了 | Alice |\n| 検索 | 進行中 | Bob |\n| 通知 | 未着手 | Diana |"},
		{userIndex: 3, body: "Go と TypeScript の例です。\n\n```go\nfunc Hello(name string) string {\n\treturn fmt.Sprintf(\"Hello, %s\", name)\n}\n```\n\n```ts\nexport const hello = (name: string) => `Hello, ${name}`;\n```\n\n```sql\nSELECT id, name FROM channel WHERE is_private = false ORDER BY name;\n```"},
		{userIndex: 4, body: longText, reactions: map[string]int{"👀": 3}},
		{userIndex: 5, body: "🎉🎉🎉"},
		{userIndex: 0, body: "<@" + users[1].ID + "> <@" + users[len(users)-len(sampleUserNames)].ID + "> レビューをお願いします。<@&" + developersGroupID + "> にも共有します。詳細は <#" + channelsByName["dev/frontend"].ID + "> と <#" + channelsByName["general"].ID + "> を見てください", reactions: map[string]int{"👍": 6, "🙏": 2, "✅": 1}},
		{userIndex: 6, body: "参考資料です https://github.com/example/repo", link: &entity.MessageLink{URL: "https://github.com/example/repo", OGP: entity.OGPData{Title: stringPtr("Example Repository"), Description: stringPtr("A sample repository for demonstration"), SiteName: stringPtr("GitHub")}}},
		{userIndex: 1, body: "最初の挨拶はここです " + permalink, link: &entity.MessageLink{URL: permalink, LinkedMessageID: &general.ID}},
		{userIndex: 7, body: "今ここにいます", location: &entity.MessageLocation{Latitude: 35.681236, Longitude: 139.767125, Label: stringPtr("東京駅")}},
		{userIndex: 2, body: "この文章はあとから編集しました（編集済みの表示）", edited: true},
		{userIndex: 3, body: "このメッセージは削除されました", deleted: true},
		{userIndex: 0, body: "リリース日の相談をスレッドでしましょう", replies: []string{"金曜はどうでしょう？", "金曜は QA が間に合わないかもしれません", "では来週の火曜で", "了解です 👍", "カレンダーに入れておきます"}},
		{userIndex: len(sampleUserNames) + 3, body: "山田です。表示名が日本語だけでも @ の候補から選べます"},
		{userIndex: 9, body: "短いメッセージ"},
	}

	at := time.Now().Add(-3 * time.Hour)
	next := func() time.Time {
		at = at.Add(4 * time.Minute)
		return at
	}
	// 以前の名前から変更したお知らせ
	if err := createSystemMessage(ctx, client, ch, entity.SystemMessageKindChannelNameChanged, users[1].ID, map[string]any{"from": "sample", "to": "showcase"}, next()); err != nil {
		return err
	}
	if err := createSystemMessage(ctx, client, ch, entity.SystemMessageKindChannelDescriptionChanged, users[1].ID, map[string]any{}, next()); err != nil {
		return err
	}
	if err := createSystemMessage(ctx, client, ch, entity.SystemMessageKindMemberAdded, users[1].ID, map[string]any{"addedBy": users[1].ID, "userId": users[3].ID}, next()); err != nil {
		return err
	}

	messageRepo := repository.NewMessageRepository(client)
	linkRepo := repository.NewLinkRepository(client)
	channelID := uuid.MustParse(ch.ID)
	for _, sample := range samples {
		author := users[sample.userIndex]
		createdAt := next()
		builder := client.Message.Create().
			SetChannelID(channelID).
			SetUserID(uuid.MustParse(author.ID)).
			SetBody(sample.body).
			SetCreatedAt(createdAt)
		if sample.edited {
			builder.SetEditedAt(createdAt.Add(2 * time.Minute))
		}
		if sample.deleted {
			builder.SetDeletedAt(createdAt.Add(time.Minute)).SetDeletedBy(uuid.MustParse(author.ID))
		}
		if loc := sample.location; loc != nil {
			builder.SetLocationLatitude(loc.Latitude).SetLocationLongitude(loc.Longitude).SetNillableLocationLabel(loc.Label)
		}
		msg, err := builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create showcase message: %w", err)
		}
		i := 0
		for emoji, count := range sample.reactions {
			for _, user := range users[:count] {
				reaction := &entity.MessageReaction{MessageID: msg.ID.String(), UserID: user.ID, Emoji: emoji, CreatedAt: createdAt.Add(time.Duration(i) * time.Second)}
				if err := messageRepo.AddReaction(ctx, reaction); err != nil {
					return fmt.Errorf("failed to create reaction: %w", err)
				}
				i++
			}
		}
		if sample.link != nil {
			sample.link.MessageID = msg.ID.String()
			sample.link.CreatedAt = createdAt
			if err := linkRepo.Create(ctx, sample.link); err != nil {
				return fmt.Errorf("failed to create link: %w", err)
			}
		}
		for j, reply := range sample.replies {
			if err := client.Message.Create().
				SetChannelID(channelID).
				SetUserID(uuid.MustParse(users[(sample.userIndex+j+1)%len(users)].ID)).
				SetParentID(msg.ID).
				SetBody(reply).
				SetCreatedAt(createdAt.Add(time.Duration(j+1) * 30 * time.Second)).
				Exec(ctx); err != nil {
				return fmt.Errorf("failed to create reply: %w", err)
			}
		}
		if len(sample.replies) > 0 {
			if err := client.MessagePin.Create().SetChannelID(channelID).SetMessageID(msg.ID).SetPinnedByID(uuid.MustParse(users[1].ID)).Exec(ctx); err != nil {
				return fmt.Errorf("failed to pin message: %w", err)
			}
			if err := createSystemMessage(ctx, client, ch, entity.SystemMessageKindMessagePinned, users[1].ID, map[string]any{"messageId": msg.ID.String(), "pinnedBy": users[1].ID}, next()); err != nil {
				return err
			}
		}
	}

	// アプリの投稿名義のボット
	bot, err := client.User.Create().
		SetEmail("deploy-bot@example.com").
		SetPasswordHash("!").
		SetDisplayName("Deploy Bot").
		SetIsApp(true).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	if err := client.App.Create().
		SetWorkspaceID(ch.WorkspaceID).
		SetName("Deploy Bot").
		SetDescription("デプロイの結果をお知らせします").
		SetPermissions([]string{string(entity.AppPermissionPostJoinedChannels), string(entity.AppPermissionPostThreadReplies)}).
		SetDefaultChannelID(channelID).
		SetCreatedByID(uuid.MustParse(users[1].ID)).
		SetBotUser(bot).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}
	if err := client.ChannelMember.Create().SetChannelID(channelID).SetUser(bot).Exec(ctx); err != nil {
		return fmt.Errorf("failed to add app to channel: %w", err)
	}
	if err := client.Message.Create().
		SetChannelID(channelID).
		SetUserID(bot.ID).
		SetBody("✅ v1.2.3 を本番環境にデプロイしました").
		SetCreatedAt(next()).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to create bot message: %w", err)
	}
	return nil
}

// createScrollTestMessages は高さのばらつくメッセージを大量に入れ、仮想スクロールと前後の読み込みを確かめられるようにします
func createScrollTestMessages(ctx context.Context, client *ent.Client, users []*entity.User, ch *entity.Channel) error {
	rng := rand.New(rand.NewPCG(1, 42))
	userIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		userIDs[i] = uuid.MustParse(user.ID)
	}
	start := time.Now().Add(-bulkSpan)
	messages := generateBulkMessages(rng, userIDs, start, scrollTestMessages)
	for i := range messages {
		m := &messages[i]
		m.body = fmt.Sprintf("No.%d %s", i+1, m.body)
		switch {
		case i%13 == 0:
			m.body += "\n\n" + strings.Repeat("複数行にわたる長めの本文です。", 12)
		case i%17 == 0:
			m.body += "\n\n```\nconst row = " + fmt.Sprint(i) + ";\nconsole.log(row);\n```"
		case i%19 == 0:
			m.body += "\n\n- 箇条書き 1\n- 箇条書き 2\n- 箇条書き 3"
		}
	}
	channelID := uuid.MustParse(ch.ID)
	for i := 0; i < len(messages); i += bulkBatchSize {
		if err := insertBulkMessages(ctx, client, channelID, messages[i:min(i+bulkBatchSize, len(messages))]); err != nil {
			return err
		}
	}
	return nil
}

// createChatter は新しいメッセージ順の並びを確かめるため、チャンネルごとに時刻をずらした会話を入れます
func createChatter(ctx context.Context, client *ent.Client, users []*entity.User, ch *entity.Channel) error {
	offset := time.Duration(len(ch.Name)) * 7 * time.Minute
	for i, body := range []string{"おはようございます", "資料を共有しました", "確認します 👀"} {
		if err := client.Message.Create().
			SetChannelID(uuid.MustParse(ch.ID)).
			SetUserID(uuid.MustParse(users[(i+len(ch.Name))%len(users)].ID)).
			SetBody(body).
			SetCreatedAt(time.Now().Add(-offset - time.Duration(3-i)*time.Minute)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create message in %s: %w", ch.Name, err)
		}
	}
	return nil
}
