package repository

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagegroupmention"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/usergroup"
	"github.com/newt239/chat/ent/usergroupmember"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

// @channel / @here を単語として含む本文に一致する POSIX 正規表現
const broadcastMentionPattern = `(^|[^[:alnum:]_])@(channel|here)([^[:alnum:]_-]|$)`

// 検索用文書は関連テーブルを配列やフラグにまとめて 1 本の SQL で読む
// $1: @channel / @here の正規表現。WHERE 句は呼び出し側で足す
const searchDocumentSQL = `
	SELECT m.id, c.channel_workspace, m.message_channel, m.message_user, m.message_parent, m.body, m.created_at,
		ARRAY(SELECT a.file_name FROM attachments a WHERE a.attachment_message = m.id ORDER BY a.created_at),
		ARRAY(SELECT a.mime_type FROM attachments a WHERE a.attachment_message = m.id),
		ARRAY(SELECT um.message_user_mention_user::text FROM message_user_mentions um WHERE um.message_user_mention_message = m.id),
		ARRAY(SELECT gm.message_group_mention_group::text FROM message_group_mentions gm WHERE gm.message_group_mention_message = m.id),
		m.body ~ $1,
		EXISTS (SELECT 1 FROM message_links l WHERE l.message_link_message = m.id),
		EXISTS (SELECT 1 FROM message_pins p WHERE p.message_pin_message = m.id),
		EXISTS (SELECT 1 FROM messages r WHERE r.message_parent = m.id AND r.deleted_at IS NULL)
	FROM messages m JOIN channels c ON c.id = m.message_channel
	WHERE m.deleted_at IS NULL AND `

func (r *messageRepository) FindSearchScope(ctx context.Context, workspaceID string, userID string) (*domainrepository.MessageSearchScope, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	viewable, err := client.Channel.Query().Where(viewableChannel(workspaceID, uid)).IDs(ctx)
	if err != nil {
		return nil, err
	}
	joined, err := client.Channel.Query().
		Where(channel.HasWorkspaceWith(workspace.ID(workspaceID)), channel.HasMembersWith(channelmember.HasUserWith(user.ID(uid)))).
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := client.UserGroup.Query().
		Where(usergroup.HasWorkspaceWith(workspace.ID(workspaceID)), usergroup.HasMembersWith(usergroupmember.HasUserWith(user.ID(uid)))).
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	return &domainrepository.MessageSearchScope{
		UserID:             userID,
		ViewableChannelIDs: uuidStrings(viewable),
		JoinedChannelIDs:   uuidStrings(joined),
		GroupIDs:           uuidStrings(groups),
	}, nil
}

func (r *messageRepository) FindSearchDocuments(ctx context.Context, messageIDs []string) ([]domainrepository.MessageSearchDocument, error) {
	if len(messageIDs) == 0 {
		return []domainrepository.MessageSearchDocument{}, nil
	}
	if _, err := parseUUIDs(messageIDs, "message ID"); err != nil {
		return nil, err
	}
	return r.querySearchDocuments(ctx, "m.id = ANY($2::uuid[])", pq.Array(messageIDs))
}

func (r *messageRepository) FindSearchDocumentsAfter(ctx context.Context, afterID string, limit int) ([]domainrepository.MessageSearchDocument, error) {
	after := uuid.Nil
	if afterID != "" {
		parsed, err := utils.ParseUUID(afterID, "message ID")
		if err != nil {
			return nil, err
		}
		after = parsed
	}
	return r.querySearchDocuments(ctx, "m.id > $2 ORDER BY m.id LIMIT $3", after, limit)
}

func (r *messageRepository) querySearchDocuments(ctx context.Context, where string, args ...any) ([]domainrepository.MessageSearchDocument, error) {
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, searchDocumentSQL+where, append([]any{broadcastMentionPattern}, args...)...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	docs := []domainrepository.MessageSearchDocument{}
	for rows.Next() {
		var (
			id, channelID, senderID                 uuid.UUID
			workspaceID, body                       string
			parentID                                uuid.NullUUID
			createdAt                               time.Time
			fileNames, mimeTypes, userIDs, groupIDs pq.StringArray
			mentionsChannel, hasLink, pinned, reply bool
		)
		if err := rows.Scan(&id, &workspaceID, &channelID, &senderID, &parentID, &body, &createdAt,
			&fileNames, &mimeTypes, &userIDs, &groupIDs, &mentionsChannel, &hasLink, &pinned, &reply); err != nil {
			return nil, err
		}
		doc := domainrepository.MessageSearchDocument{
			ID:                id.String(),
			WorkspaceID:       workspaceID,
			ChannelID:         channelID.String(),
			SenderID:          senderID.String(),
			Body:              body,
			AttachmentNames:   fileNames,
			Has:               contentKinds(body, mimeTypes, hasLink),
			MentionedUserIDs:  userIDs,
			MentionedGroupIDs: groupIDs,
			MentionsChannel:   mentionsChannel,
			Pinned:            pinned,
			HasReplies:        reply,
			CreatedAt:         createdAt,
		}
		if parentID.Valid {
			pid := parentID.UUID.String()
			doc.ParentID = &pid
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

// contentKinds は添付の種類と、リンクを含むかどうかから絞り込み用の種類を求めます
func contentKinds(body string, mimeTypes []string, hasLink bool) []domainrepository.MessageContentKind {
	seen := map[domainrepository.MessageContentKind]bool{}
	for _, mimeType := range mimeTypes {
		switch {
		case strings.HasPrefix(mimeType, "image/"):
			seen[domainrepository.MessageContentImage] = true
		case strings.HasPrefix(mimeType, "video/"):
			seen[domainrepository.MessageContentVideo] = true
		default:
			seen[domainrepository.MessageContentFile] = true
		}
	}
	if hasLink || strings.Contains(body, "http://") || strings.Contains(body, "https://") {
		seen[domainrepository.MessageContentLink] = true
	}
	kinds := []domainrepository.MessageContentKind{}
	for _, kind := range []domainrepository.MessageContentKind{
		domainrepository.MessageContentImage, domainrepository.MessageContentVideo,
		domainrepository.MessageContentFile, domainrepository.MessageContentLink,
	} {
		if seen[kind] {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

func uuidStrings(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

func (r *messageRepository) FindMentions(ctx context.Context, input domainrepository.FindMentionsInput) ([]*entity.Message, error) {
	userID, err := utils.ParseUUID(input.UserID, "user ID")
	if err != nil {
		return nil, err
	}

	preds := []predicate.Message{
		message.DeletedAtIsNil(),
		message.HasChannelWith(viewableChannel(input.WorkspaceID, userID)),
		message.UserIDNEQ(userID),
		mentionsUser(userID),
	}
	if input.Cursor != nil {
		cursorID, err := utils.ParseUUID(input.Cursor.MessageID, "cursor message ID")
		if err != nil {
			return nil, err
		}
		preds = append(preds, message.Or(
			message.CreatedAtLT(input.Cursor.CreatedAt),
			message.And(message.CreatedAtEQ(input.Cursor.CreatedAt), message.IDLT(cursorID)),
		))
	}

	client := transaction.ResolveClient(ctx, r.client)
	messages, err := client.Message.Query().
		Where(preds...).
		Order(ent.Desc(message.FieldCreatedAt), ent.Desc(message.FieldID)).
		Limit(input.Limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toMessageEntities(messages), nil
}

// viewableChannel はワークスペース内の公開チャンネルと、参加している非公開チャンネル（DM を含む）に一致します
func viewableChannel(workspaceID string, userID uuid.UUID) predicate.Channel {
	return channel.And(
		channel.HasWorkspaceWith(workspace.ID(workspaceID)),
		channel.Or(
			channel.IsPrivate(false),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(userID))),
		),
	)
}

// mentionsUser は本人へのメンション、所属グループへのメンション、参加チャンネルでの @channel / @here に一致します
func mentionsUser(userID uuid.UUID) predicate.Message {
	return message.Or(
		message.HasUserMentionsWith(messageusermention.HasUserWith(user.ID(userID))),
		message.HasGroupMentionsWith(messagegroupmention.HasGroupWith(
			usergroup.HasMembersWith(usergroupmember.HasUserWith(user.ID(userID))),
		)),
		message.And(
			message.HasChannelWith(channel.HasMembersWith(channelmember.HasUserWith(user.ID(userID)))),
			predicate.Message(func(s *sql.Selector) {
				s.Where(sql.P(func(b *sql.Builder) {
					b.WriteString(s.C(message.FieldBody)).WriteString(" ~ ").Arg(broadcastMentionPattern)
				}))
			}),
		),
	)
}
