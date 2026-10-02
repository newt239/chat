package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelreadstate"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/ent/workspacemember"
)

const (
	bulkBatchSize = 2000
	bulkSpan      = 180 * 24 * time.Hour
)

var bulkWords = []string{
	"設計", "レビュー", "リリース", "障害", "対応", "確認", "お願いします", "資料", "会議", "議事録",
	"deploy", "review", "design", "bug", "fix", "release", "meeting", "schedule", "検索", "パフォーマンス",
}

type bulkMessage struct {
	id        uuid.UUID
	userID    uuid.UUID
	parentID  *uuid.UUID
	body      string
	createdAt time.Time
	deleted   bool
	mentionee *uuid.UUID
	// 本文に <@channel> を含む
	mentionsChannel bool
}

// BulkMessages は channelCount 個のチャンネルを足したうえで、全チャンネルに perChannel 件ずつ
// 返信・削除済み・メンションを混ぜたメッセージと既読位置を投入します
func BulkMessages(ctx context.Context, client *ent.Client, channelCount int, perChannel int) error {
	if err := createBulkChannels(ctx, client, channelCount); err != nil {
		return err
	}
	channels, err := client.Channel.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to load channels: %w", err)
	}
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 42))
	start := time.Now().Add(-bulkSpan)
	for _, ch := range channels {
		userIDs, err := client.WorkspaceMember.Query().
			Where(workspacemember.WorkspaceID(ch.WorkspaceID)).
			QueryUser().IDs(ctx)
		if err != nil {
			return fmt.Errorf("failed to load members: %w", err)
		}
		if len(userIDs) == 0 {
			continue
		}
		messages := generateBulkMessages(rng, userIDs, start, perChannel)
		for i := 0; i < len(messages); i += bulkBatchSize {
			if err := insertBulkMessages(ctx, client, ch.ID, messages[i:min(i+bulkBatchSize, len(messages))]); err != nil {
				return err
			}
		}
		if err := insertBulkReadStates(ctx, client, rng, ch.ID, userIDs, start); err != nil {
			return err
		}
	}
	return nil
}

// createBulkChannels は最初のワークスペースに全メンバーが参加する公開チャンネルを作ります
func createBulkChannels(ctx context.Context, client *ent.Client, n int) error {
	if n == 0 {
		return nil
	}
	ws, err := client.Workspace.Query().Order(ent.Asc(workspace.FieldCreatedAt)).First(ctx)
	if err != nil {
		return fmt.Errorf("failed to load workspace: %w", err)
	}
	userIDs, err := ws.QueryMembers().QueryUser().IDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load members: %w", err)
	}
	suffix := uuid.NewString()[:8]
	for i := range n {
		ch, err := client.Channel.Create().
			SetName(fmt.Sprintf("bench-%s-%03d", suffix, i)).
			SetWorkspace(ws).
			SetCreatedByID(userIDs[0]).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		members := make([]*ent.ChannelMemberCreate, 0, len(userIDs))
		for _, uid := range userIDs {
			members = append(members, client.ChannelMember.Create().SetChannel(ch).SetUserID(uid))
		}
		if err := client.ChannelMember.CreateBulk(members...).Exec(ctx); err != nil {
			return fmt.Errorf("failed to add members: %w", err)
		}
	}
	return nil
}

// generateBulkMessages は古い順に並んだメッセージを作ります。返信は同じチャンネルのより古い親を指す
func generateBulkMessages(rng *rand.Rand, userIDs []uuid.UUID, start time.Time, n int) []bulkMessage {
	step := bulkSpan / time.Duration(n+1)
	messages := make([]bulkMessage, 0, n)
	parents := []uuid.UUID{}
	for i := range n {
		m := bulkMessage{
			id:        uuid.New(),
			userID:    userIDs[rng.IntN(len(userIDs))],
			body:      bulkBody(rng),
			createdAt: start.Add(step * time.Duration(i+1)),
			deleted:   rng.IntN(100) < 3,
		}
		if len(parents) > 0 && rng.IntN(100) < 15 {
			parent := parents[max(len(parents)-1-rng.IntN(20), 0)]
			m.parentID = &parent
		} else {
			parents = append(parents, m.id)
		}
		switch r := rng.IntN(100); {
		case r < 5:
			mentionee := userIDs[rng.IntN(len(userIDs))]
			m.mentionee = &mentionee
			m.body = "<@" + mentionee.String() + "> " + m.body
		case r < 7:
			m.body = "<@channel> " + m.body
			m.mentionsChannel = true
		case r < 10:
			m.body += " https://example.com/docs"
		}
		messages = append(messages, m)
	}
	return messages
}

func insertBulkMessages(ctx context.Context, client *ent.Client, channelID uuid.UUID, messages []bulkMessage) error {
	builders := make([]*ent.MessageCreate, 0, len(messages))
	mentions := []*ent.MessageUserMentionCreate{}
	for _, m := range messages {
		b := client.Message.Create().
			SetID(m.id).
			SetChannelID(channelID).
			SetUserID(m.userID).
			SetNillableParentID(m.parentID).
			SetBody(m.body).
			SetMentionsChannel(m.mentionsChannel).
			SetCreatedAt(m.createdAt)
		if m.deleted {
			b.SetDeletedAt(m.createdAt.Add(time.Minute)).SetDeletedBy(m.userID)
		}
		builders = append(builders, b)
		if m.mentionee != nil {
			mentions = append(mentions, client.MessageUserMention.Create().SetMessageID(m.id).SetUserID(*m.mentionee))
		}
	}
	if err := client.Message.CreateBulk(builders...).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert messages: %w", err)
	}
	if err := client.MessageUserMention.CreateBulk(mentions...).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert mentions: %w", err)
	}
	return nil
}

// insertBulkReadStates は既読位置のないメンバーに、期間内のランダムな既読位置を作ります
func insertBulkReadStates(ctx context.Context, client *ent.Client, rng *rand.Rand, channelID uuid.UUID, userIDs []uuid.UUID, start time.Time) error {
	existing, err := client.ChannelReadState.Query().
		Where(channelreadstate.ChannelID(channelID)).
		QueryUser().IDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load read states: %w", err)
	}
	builders := []*ent.ChannelReadStateCreate{}
	for _, uid := range userIDs {
		if slices.Contains(existing, uid) {
			continue
		}
		lastReadAt := start.Add(time.Duration(rng.Int64N(int64(bulkSpan))))
		builders = append(builders, client.ChannelReadState.Create().SetChannelID(channelID).SetUserID(uid).SetLastReadAt(lastReadAt))
	}
	if err := client.ChannelReadState.CreateBulk(builders...).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert read states: %w", err)
	}
	return nil
}

func bulkBody(rng *rand.Rand) string {
	words := make([]string, 3+rng.IntN(8))
	for i := range words {
		words[i] = bulkWords[rng.IntN(len(bulkWords))]
	}
	return strings.Join(words, " ")
}
