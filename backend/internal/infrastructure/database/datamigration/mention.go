package datamigration

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lib/pq"
)

var (
	legacyMentionPattern = regexp.MustCompile(`@([A-Za-z0-9_-]+)`)
	legacyChannelPattern = regexp.MustCompile(`#([A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*)`)
)

type named struct {
	id   string
	name string
}

// legacyContext は 1 件のメッセージを書き換えるのに使う、投稿時に保存したメンションとワークスペースのチャンネルです
type legacyContext struct {
	users    []named
	groups   []named
	channels map[string]string
}

// rewriteLegacyBody は名前で書かれたメンションとチャンネルを ID 記法に置き換えます。
// ユーザーは保存済みメンションの表示名との前方一致、グループは名前の完全一致で、以前の表示と同じ規則で解決します
func rewriteLegacyBody(body string, lc legacyContext) (rewritten string, mentionsChannel, mentionsHere bool) {
	// バッククォートで区切った奇数番目はコードなので書き換えない
	segments := strings.Split(body, "`")
	for i := 0; i < len(segments); i += 2 {
		segments[i] = replaceAfterBoundary(segments[i], legacyMentionPattern, func(token string) (string, bool) {
			switch token {
			case "channel":
				mentionsChannel = true
				return "<@channel>", true
			case "here":
				mentionsHere = true
				return "<@here>", true
			}
			lower := strings.ToLower(token)
			for _, u := range lc.users {
				if strings.HasPrefix(strings.ToLower(u.name), lower) {
					return "<@" + u.id + ">", true
				}
			}
			for _, g := range lc.groups {
				if g.name == token {
					return "<@&" + g.id + ">", true
				}
			}
			return "", false
		})
		segments[i] = replaceAfterBoundary(segments[i], legacyChannelPattern, func(name string) (string, bool) {
			id, ok := lc.channels[name]
			return "<#" + id + ">", ok
		})
	}
	return strings.Join(segments, "`"), mentionsChannel, mentionsHere
}

// replaceAfterBoundary は単語や URL の途中（メールアドレスや URL のフラグメント）を除いて置き換えます
func replaceAfterBoundary(text string, pattern *regexp.Regexp, replace func(token string) (string, bool)) string {
	var b strings.Builder
	last := 0
	for _, m := range pattern.FindAllStringSubmatchIndex(text, -1) {
		if prev, _ := utf8.DecodeLastRuneInString(text[:m[0]]); m[0] > 0 && !unicode.IsSpace(prev) && !strings.ContainsRune("([（「『、。，,!！?？", prev) {
			continue
		}
		replacement, ok := replace(text[m[2]:m[3]])
		if !ok {
			continue
		}
		b.WriteString(text[last:m[0]])
		b.WriteString(replacement)
		last = m[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

// convertLegacyMentions は名前で書かれた既存メッセージの本文を ID 記法に書き換え、グループへのメンションを今のメンバーで展開して保存します
func convertLegacyMentions(ctx context.Context, tx *sql.Tx) error {
	channels, err := loadChannelNames(ctx, tx)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT m.id, m.body, c.channel_workspace,
			ARRAY(SELECT u.id::text || ':' || u.display_name FROM message_user_mentions um JOIN users u ON u.id = um.message_user_mention_user
				WHERE um.message_user_mention_message = m.id ORDER BY um.created_at),
			ARRAY(SELECT g.id::text || ':' || g.name FROM message_group_mentions gm JOIN user_groups g ON g.id = gm.message_group_mention_group
				WHERE gm.message_group_mention_message = m.id)
		FROM messages m JOIN channels c ON c.id = m.message_channel
		WHERE m.body ~ '[@#]'`)
	if err != nil {
		return err
	}
	type legacyMessage struct {
		id, body string
		lc       legacyContext
	}
	var messages []legacyMessage
	for rows.Next() {
		var id, body, workspaceID string
		var users, groups pq.StringArray
		if err := rows.Scan(&id, &body, &workspaceID, &users, &groups); err != nil {
			_ = rows.Close()
			return err
		}
		messages = append(messages, legacyMessage{id: id, body: body, lc: legacyContext{users: splitNamed(users), groups: splitNamed(groups), channels: channels[workspaceID]}})
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, m := range messages {
		body, mentionsChannel, mentionsHere := rewriteLegacyBody(m.body, m.lc)
		if body == m.body {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE messages SET body = $2, mentions_channel = $3, mentions_here = $4 WHERE id = $1`,
			m.id, body, mentionsChannel, mentionsHere); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO message_user_mentions (id, created_at, message_user_mention_message, message_user_mention_user, via_group_id)
		SELECT DISTINCT ON (gm.message_group_mention_message, ugm.user_group_member_user)
			gen_random_uuid(), gm.created_at, gm.message_group_mention_message, ugm.user_group_member_user, gm.message_group_mention_group
		FROM message_group_mentions gm
		JOIN user_group_members ugm ON ugm.user_group_member_group = gm.message_group_mention_group
		WHERE NOT EXISTS (
			SELECT 1 FROM message_user_mentions um
			WHERE um.message_user_mention_message = gm.message_group_mention_message AND um.message_user_mention_user = ugm.user_group_member_user
		)`)
	return err
}

// loadChannelNames はワークスペースごとに、チャンネル名から ID を引ける表を作ります。DM は名前で参照しないため含めない
func loadChannelNames(ctx context.Context, tx *sql.Tx) (map[string]map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, name, channel_workspace FROM channels WHERE channel_type IN ('public', 'private')`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := map[string]map[string]string{}
	for rows.Next() {
		var id, name, workspaceID string
		if err := rows.Scan(&id, &name, &workspaceID); err != nil {
			return nil, err
		}
		if result[workspaceID] == nil {
			result[workspaceID] = map[string]string{}
		}
		result[workspaceID][name] = id
	}
	return result, rows.Err()
}

func splitNamed(values []string) []named {
	result := make([]named, 0, len(values))
	for _, v := range values {
		id, name, _ := strings.Cut(v, ":")
		result = append(result, named{id: id, name: name})
	}
	return result
}
