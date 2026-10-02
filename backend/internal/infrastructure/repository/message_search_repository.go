package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

// 検索用文書は関連テーブルを配列やフラグにまとめて 1 本の SQL で読む。WHERE 句は呼び出し側で足す
const searchDocumentSQL = `
	SELECT m.id, c.workspace_id, m.channel_id, m.user_id, m.parent_id, m.body, m.created_at,
		ARRAY(SELECT a.file_name FROM attachment a WHERE a.message_id = m.id ORDER BY a.created_at),
		ARRAY(SELECT a.mime_type FROM attachment a WHERE a.message_id = m.id),
		ARRAY(SELECT um.user_id::text FROM message_user_mention um WHERE um.message_id = m.id),
		ARRAY(SELECT gm.group_id::text FROM message_group_mention gm WHERE gm.message_id = m.id),
		m.mentions_channel OR m.mentions_here,
		EXISTS (SELECT 1 FROM message_link l WHERE l.message_id = m.id),
		EXISTS (SELECT 1 FROM message_pin p WHERE p.message_id = m.id),
		EXISTS (SELECT 1 FROM message r WHERE r.parent_id = m.id AND r.deleted_at IS NULL),
		m.location_latitude IS NOT NULL
	FROM message m JOIN channel c ON c.id = m.channel_id
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
		Where(channel.WorkspaceID(workspaceID), channel.HasMembersWith(channelmember.UserID(uid))).
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	return &domainrepository.MessageSearchScope{
		UserID:             userID,
		ViewableChannelIDs: uuidStrings(viewable),
		JoinedChannelIDs:   uuidStrings(joined),
	}, nil
}

func (r *messageRepository) FindSearchDocuments(ctx context.Context, messageIDs []string) ([]domainrepository.MessageSearchDocument, error) {
	if len(messageIDs) == 0 {
		return []domainrepository.MessageSearchDocument{}, nil
	}
	if _, err := parseUUIDs(messageIDs, "message ID"); err != nil {
		return nil, err
	}
	return r.querySearchDocuments(ctx, "m.id = ANY($1::uuid[])", pq.Array(messageIDs))
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
	return r.querySearchDocuments(ctx, "m.id > $1 ORDER BY m.id LIMIT $2", after, limit)
}

func (r *messageRepository) querySearchDocuments(ctx context.Context, where string, args ...any) ([]domainrepository.MessageSearchDocument, error) {
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, searchDocumentSQL+where, args...)
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
			hasLocation                             bool
		)
		if err := rows.Scan(&id, &workspaceID, &channelID, &senderID, &parentID, &body, &createdAt,
			&fileNames, &mimeTypes, &userIDs, &groupIDs, &mentionsChannel, &hasLink, &pinned, &reply, &hasLocation); err != nil {
			return nil, err
		}
		doc := domainrepository.MessageSearchDocument{
			ID:                id.String(),
			WorkspaceID:       workspaceID,
			ChannelID:         channelID.String(),
			SenderID:          senderID.String(),
			Body:              body,
			AttachmentNames:   fileNames,
			Has:               contentKinds(body, mimeTypes, hasLink, hasLocation),
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

// contentKinds は添付の種類と、リンク・位置情報を含むかどうかから絞り込み用の種類を求めます
func contentKinds(body string, mimeTypes []string, hasLink, hasLocation bool) []domainrepository.MessageContentKind {
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
	seen[domainrepository.MessageContentLocation] = hasLocation
	kinds := []domainrepository.MessageContentKind{}
	for _, kind := range []domainrepository.MessageContentKind{
		domainrepository.MessageContentImage, domainrepository.MessageContentVideo,
		domainrepository.MessageContentFile, domainrepository.MessageContentLink,
		domainrepository.MessageContentLocation,
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
		channel.WorkspaceID(workspaceID),
		channel.Or(
			channel.IsPrivate(false),
			channel.HasMembersWith(channelmember.UserID(userID)),
		),
	)
}

// mentionsUser は本人へのメンション（グループ経由を含む）と、参加チャンネルでの @channel / @here に一致します
func mentionsUser(userID uuid.UUID) predicate.Message {
	return message.Or(
		message.HasUserMentionsWith(messageusermention.UserID(userID)),
		message.And(
			message.HasChannelWith(channel.HasMembersWith(channelmember.UserID(userID))),
			message.Or(message.MentionsChannel(true), message.MentionsHere(true)),
		),
	)
}
