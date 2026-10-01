package message

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

const previewExcerptRunes = 200

// MessageOutputBuilder はメッセージエンティティと関連データから閲覧者向けの MessageOutput を組み立てます
type MessageOutputBuilder struct {
	messageRepo      domainrepository.MessageRepository
	userRepo         domainrepository.UserRepository
	userGroupRepo    domainrepository.UserGroupRepository
	userMentionRepo  domainrepository.MessageUserMentionRepository
	groupMentionRepo domainrepository.MessageGroupMentionRepository
	linkRepo         domainrepository.MessageLinkRepository
	attachmentRepo   domainrepository.AttachmentRepository
	pinRepo          domainrepository.PinRepository
	pollRepo         domainrepository.PollRepository
	channelAccessSvc service.ChannelAccessService
}

func NewMessageOutputBuilder(
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	pinRepo domainrepository.PinRepository,
	pollRepo domainrepository.PollRepository,
	channelAccessSvc service.ChannelAccessService,
) *MessageOutputBuilder {
	return &MessageOutputBuilder{
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		userGroupRepo:    userGroupRepo,
		userMentionRepo:  userMentionRepo,
		groupMentionRepo: groupMentionRepo,
		linkRepo:         linkRepo,
		attachmentRepo:   attachmentRepo,
		pinRepo:          pinRepo,
		pollRepo:         pollRepo,
		channelAccessSvc: channelAccessSvc,
	}
}

type relatedData struct {
	userMentions  map[string][]*entity.MessageUserMention
	groupMentions map[string][]*entity.MessageGroupMention
	links         map[string][]*entity.MessageLink
	reactions     map[string][]*entity.MessageReaction
	attachments   map[string][]*entity.Attachment
	pins          map[string]*entity.MessagePin
	groups        map[string]*entity.UserGroup
	polls         map[string]*entity.Poll
	pollVotes     map[string][]*entity.PollVote
}

// Build は viewerID から見た MessageOutput を組み立てます。メッセージリンクの引用は viewerID が参照できるものだけ含めます
func (b *MessageOutputBuilder) Build(ctx context.Context, viewerID string, messages []*entity.Message) ([]MessageOutput, error) {
	if len(messages) == 0 {
		return []MessageOutput{}, nil
	}

	messageIDs := make([]string, len(messages))
	for idx, msg := range messages {
		messageIDs[idx] = msg.ID
	}

	related, err := b.fetchRelatedData(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	linkedIDs := make([]string, 0)
	for _, links := range related.links {
		for _, link := range links {
			if link.LinkedMessageID != nil {
				linkedIDs = append(linkedIDs, *link.LinkedMessageID)
			}
		}
	}
	linked, channels, err := b.fetchAccessibleMessages(ctx, viewerID, linkedIDs)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0)
	for _, msg := range messages {
		userIDs = append(userIDs, msg.UserID)
		if msg.DeletedBy != nil {
			userIDs = append(userIDs, *msg.DeletedBy)
		}
	}
	for _, msg := range linked {
		userIDs = append(userIDs, msg.UserID)
	}
	for _, reactions := range related.reactions {
		for _, reaction := range reactions {
			userIDs = append(userIDs, reaction.UserID)
		}
	}
	for _, pin := range related.pins {
		userIDs = append(userIDs, pin.PinnedBy)
	}
	users, err := b.fetchUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	previews := make(map[string]*MessagePreviewOutput, len(linked))
	for _, msg := range linked {
		previews[msg.ID] = buildPreview(msg, channels[msg.ChannelID], users)
	}

	outputs := make([]MessageOutput, 0, len(messages))
	for _, msg := range messages {
		output := assemble(msg, related, previews, users)
		if poll := related.polls[msg.ID]; poll != nil {
			output.Poll = buildPollOutput(poll, related.pollVotes[poll.ID], viewerID, time.Now())
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

// BuildPreview は viewerID が参照できるメッセージの引用カードを返します。参照できなければ ErrMessageNotFound を返します
func (b *MessageOutputBuilder) BuildPreview(ctx context.Context, viewerID, messageID string) (*MessagePreviewOutput, error) {
	linked, channels, err := b.fetchAccessibleMessages(ctx, viewerID, []string{messageID})
	if err != nil {
		return nil, err
	}
	if len(linked) == 0 {
		return nil, ErrMessageNotFound
	}
	users, err := b.fetchUsers(ctx, []string{linked[0].UserID})
	if err != nil {
		return nil, err
	}
	return buildPreview(linked[0], channels[linked[0].ChannelID], users), nil
}

// fetchAccessibleMessages は削除されておらず viewerID が参照できるメッセージと、そのチャンネルを返します
func (b *MessageOutputBuilder) fetchAccessibleMessages(ctx context.Context, viewerID string, ids []string) ([]*entity.Message, map[string]*entity.Channel, error) {
	channels := map[string]*entity.Channel{}
	if len(ids) == 0 {
		return nil, channels, nil
	}

	messages, err := b.messageRepo.FindByIDs(ctx, uniqueStrings(ids))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch linked messages: %w", err)
	}

	accessible := make([]*entity.Message, 0, len(messages))
	for _, msg := range messages {
		if msg.DeletedAt != nil {
			continue
		}
		ch, checked := channels[msg.ChannelID]
		if !checked {
			ch, err = b.channelAccessSvc.EnsureChannelAccess(ctx, msg.ChannelID, viewerID)
			if err != nil && !errors.Is(err, domainerrors.ErrUnauthorized) && !errors.Is(err, domainerrors.ErrChannelNotFound) {
				return nil, nil, err
			}
			channels[msg.ChannelID] = ch
		}
		if ch != nil {
			accessible = append(accessible, msg)
		}
	}
	return accessible, channels, nil
}

func (b *MessageOutputBuilder) fetchRelatedData(ctx context.Context, messageIDs []string) (*relatedData, error) {
	userMentions, err := b.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user mentions: %w", err)
	}
	groupMentions, err := b.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group mentions: %w", err)
	}
	links, err := b.linkRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}
	reactions, err := b.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}
	attachments, err := b.attachmentRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attachments: %w", err)
	}
	pins, err := b.pinRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pins: %w", err)
	}

	polls, err := b.pollRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch polls: %w", err)
	}
	pollIDs := make([]string, 0, len(polls))
	for _, poll := range polls {
		pollIDs = append(pollIDs, poll.ID)
	}
	votes, err := b.pollRepo.FindVotesByPollIDs(ctx, pollIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch poll votes: %w", err)
	}

	groupIDs := make([]string, 0, len(groupMentions))
	for _, mention := range groupMentions {
		groupIDs = append(groupIDs, mention.GroupID)
	}
	groups := map[string]*entity.UserGroup{}
	if len(groupIDs) > 0 {
		groupList, err := b.userGroupRepo.FindByIDs(ctx, uniqueStrings(groupIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch groups: %w", err)
		}
		for _, group := range groupList {
			groups[group.ID] = group
		}
	}

	return &relatedData{
		userMentions:  groupByMessageID(userMentions, func(m *entity.MessageUserMention) string { return m.MessageID }),
		groupMentions: groupByMessageID(groupMentions, func(m *entity.MessageGroupMention) string { return m.MessageID }),
		links:         groupByMessageID(links, func(l *entity.MessageLink) string { return l.MessageID }),
		reactions:     reactions,
		attachments:   attachments,
		pins:          pins,
		groups:        groups,
		polls:         polls,
		pollVotes:     groupByMessageID(votes, func(v *entity.PollVote) string { return v.PollID }),
	}, nil
}

func (b *MessageOutputBuilder) fetchUsers(ctx context.Context, userIDs []string) (map[string]*entity.User, error) {
	users := map[string]*entity.User{}
	if len(userIDs) == 0 {
		return users, nil
	}
	found, err := b.userRepo.FindByIDs(ctx, uniqueStrings(userIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	for _, u := range found {
		users[u.ID] = u
	}
	return users, nil
}

func assemble(msg *entity.Message, related *relatedData, previews map[string]*MessagePreviewOutput, users map[string]*entity.User) MessageOutput {
	output := MessageOutput{
		ID:          msg.ID,
		ChannelID:   msg.ChannelID,
		UserID:      msg.UserID,
		User:        authorInfo(msg, users),
		ParentID:    msg.ParentID,
		Body:        msg.Body,
		Mentions:    make([]UserMention, 0, len(related.userMentions[msg.ID])),
		Groups:      make([]GroupMention, 0, len(related.groupMentions[msg.ID])),
		Links:       make([]LinkInfo, 0, len(related.links[msg.ID])),
		Reactions:   make([]ReactionInfo, 0, len(related.reactions[msg.ID])),
		Attachments: make([]AttachmentInfo, 0, len(related.attachments[msg.ID])),
		CreatedAt:   msg.CreatedAt,
		EditedAt:    msg.EditedAt,
		DeletedAt:   msg.DeletedAt,
		IsDeleted:   msg.DeletedAt != nil,
		Location:    msg.Location,

		MentionsChannel: msg.MentionsChannel,
		MentionsHere:    msg.MentionsHere,
		IsOfficial:      users[msg.UserID] != nil && users[msg.UserID].IsOfficial,
	}

	for _, mention := range related.userMentions[msg.ID] {
		output.Mentions = append(output.Mentions, UserMention{UserID: mention.UserID, ViaGroupID: mention.ViaGroupID})
	}
	for _, mention := range related.groupMentions[msg.ID] {
		name := ""
		if group := related.groups[mention.GroupID]; group != nil {
			name = group.Name
		}
		output.Groups = append(output.Groups, GroupMention{GroupID: mention.GroupID, Name: name})
	}
	for _, link := range related.links[msg.ID] {
		info := LinkInfo{ID: link.ID, URL: link.URL, OGP: link.OGP, LinkedMessageID: link.LinkedMessageID}
		if link.LinkedMessageID != nil {
			info.MessagePreview = previews[*link.LinkedMessageID]
		}
		output.Links = append(output.Links, info)
	}
	for _, reaction := range related.reactions[msg.ID] {
		output.Reactions = append(output.Reactions, ReactionInfo{
			User:      toUserInfo(reaction.UserID, users),
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}
	for _, attachment := range related.attachments[msg.ID] {
		output.Attachments = append(output.Attachments, AttachmentInfo{
			ID:        attachment.ID,
			FileName:  attachment.FileName,
			MimeType:  attachment.MimeType,
			SizeBytes: attachment.SizeBytes,
			Media:     attachment.Media,
		})
	}
	if msg.DeletedBy != nil {
		deletedBy := toUserInfo(*msg.DeletedBy, users)
		output.DeletedBy = &deletedBy
	}
	if pin := related.pins[msg.ID]; pin != nil {
		output.Pin = &PinInfo{PinnedBy: toUserInfo(pin.PinnedBy, users), PinnedAt: pin.PinnedAt}
	}
	return output
}

func buildPreview(msg *entity.Message, ch *entity.Channel, users map[string]*entity.User) *MessagePreviewOutput {
	return &MessagePreviewOutput{
		MessageID:   msg.ID,
		ChannelID:   msg.ChannelID,
		ChannelName: ch.Name,
		ParentID:    msg.ParentID,
		User:        authorInfo(msg, users),
		BodyExcerpt: excerpt(msg.Body, previewExcerptRunes),
		CreatedAt:   msg.CreatedAt,
	}
}

func excerpt(body string, maxRunes int) string {
	runes := []rune(body)
	if len(runes) <= maxRunes {
		return body
	}
	return string(runes[:maxRunes]) + "…"
}

func toUserInfo(userID string, users map[string]*entity.User) UserInfo {
	if u := users[userID]; u != nil {
		return UserInfo{ID: u.ID, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL, IsBot: u.IsBot}
	}
	return UserInfo{ID: userID, DisplayName: "Unknown User"}
}

// authorInfo は投稿者の情報に、アプリが投稿ごとに指定した表示名とアイコンを反映します
func authorInfo(msg *entity.Message, users map[string]*entity.User) UserInfo {
	info := toUserInfo(msg.UserID, users)
	if msg.SenderName != nil {
		info.DisplayName = *msg.SenderName
	}
	if msg.SenderAvatarURL != nil {
		info.AvatarURL = msg.SenderAvatarURL
	}
	return info
}

func groupByMessageID[T any](items []T, messageID func(T) string) map[string][]T {
	grouped := make(map[string][]T)
	for _, item := range items {
		grouped[messageID(item)] = append(grouped[messageID(item)], item)
	}
	return grouped
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}
