package repository

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/attachment"
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

func (r *messageRepository) SearchMessages(ctx context.Context, c domainrepository.MessageSearchCriteria) ([]*entity.Message, int, error) {
	viewerID, err := utils.ParseUUID(c.ViewerID, "viewer ID")
	if err != nil {
		return nil, 0, err
	}
	channelIDs, err := parseUUIDs(c.ChannelIDs, "channel ID")
	if err != nil {
		return nil, 0, err
	}
	authorIDs, err := parseUUIDs(c.AuthorIDs, "author ID")
	if err != nil {
		return nil, 0, err
	}

	preds := []predicate.Message{
		message.DeletedAtIsNil(),
		message.HasChannelWith(viewableChannel(c.WorkspaceID, viewerID)),
	}
	for _, term := range c.Terms {
		preds = append(preds, message.Or(
			message.BodyContainsFold(term),
			message.HasAttachmentsWith(attachment.FileNameContainsFold(term)),
		))
	}
	if len(channelIDs) > 0 {
		preds = append(preds, message.HasChannelWith(channel.IDIn(channelIDs...)))
	}
	if len(authorIDs) > 0 {
		preds = append(preds, message.HasUserWith(user.IDIn(authorIDs...)))
	}
	for _, kind := range c.Has {
		pred, err := hasContent(kind)
		if err != nil {
			return nil, 0, err
		}
		preds = append(preds, pred)
	}
	if c.PinnedOnly {
		preds = append(preds, message.HasPins())
	}
	if c.ThreadOnly {
		preds = append(preds, message.Or(message.HasParent(), message.HasReplies()))
	}
	if c.ExcludeReplies {
		preds = append(preds, message.Not(message.HasParent()))
	}
	if c.MentionsViewer {
		preds = append(preds, mentionsUser(viewerID))
	}
	if c.After != nil {
		preds = append(preds, message.CreatedAtGTE(*c.After))
	}
	if c.Before != nil {
		preds = append(preds, message.CreatedAtLT(*c.Before))
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Message.Query().Where(preds...)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if c.Sort == domainrepository.MessageSearchSortRelevance && len(c.Terms) > 0 {
		query = query.Order(relevanceOrder(c.Terms))
	}
	messages, err := query.
		Order(ent.Desc(message.FieldCreatedAt), ent.Desc(message.FieldID)).
		Offset(c.Offset).
		Limit(c.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return toMessageEntities(messages), total, nil
}

func (r *messageRepository) FindMentions(ctx context.Context, input domainrepository.FindMentionsInput) ([]*entity.Message, error) {
	userID, err := utils.ParseUUID(input.UserID, "user ID")
	if err != nil {
		return nil, err
	}

	preds := []predicate.Message{
		message.DeletedAtIsNil(),
		message.HasChannelWith(viewableChannel(input.WorkspaceID, userID)),
		message.Not(message.HasUserWith(user.ID(userID))),
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

func hasContent(kind domainrepository.MessageContentKind) (predicate.Message, error) {
	switch kind {
	case domainrepository.MessageContentImage:
		return message.HasAttachmentsWith(attachment.MimeTypeHasPrefix("image/")), nil
	case domainrepository.MessageContentVideo:
		return message.HasAttachmentsWith(attachment.MimeTypeHasPrefix("video/")), nil
	case domainrepository.MessageContentFile:
		return message.HasAttachmentsWith(
			attachment.Not(attachment.MimeTypeHasPrefix("image/")),
			attachment.Not(attachment.MimeTypeHasPrefix("video/")),
		), nil
	case domainrepository.MessageContentLink:
		return message.Or(
			message.HasLinks(),
			message.BodyContains("http://"),
			message.BodyContains("https://"),
		), nil
	default:
		return nil, fmt.Errorf("unknown content kind: %s", kind)
	}
}

// relevanceOrder は語をそのまま並べた句を含むものを先頭に、次に語の出現回数の合計が多い順に並べます
// ORDER BY 句ではプレースホルダの番号が振られないため、語はリテラルとしてエスケープして埋め込む
func relevanceOrder(terms []string) func(*sql.Selector) {
	return func(s *sql.Selector) {
		body := "lower(" + s.C(message.FieldBody) + ")"
		if len(terms) > 1 {
			phrase := pq.QuoteLiteral(strings.ToLower(strings.Join(terms, " ")))
			s.OrderExpr(sql.Expr(fmt.Sprintf("strpos(%s, %s) > 0 DESC", body, phrase)))
		}
		counts := make([]string, 0, len(terms))
		for _, term := range terms {
			lowered := strings.ToLower(term)
			counts = append(counts, fmt.Sprintf("(char_length(%[1]s) - char_length(replace(%[1]s, %[2]s, ''))) / %[3]d",
				body, pq.QuoteLiteral(lowered), utf8.RuneCountInString(lowered)))
		}
		s.OrderExpr(sql.Expr("(" + strings.Join(counts, " + ") + ") DESC"))
	}
}
