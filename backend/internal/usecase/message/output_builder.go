package message

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// MessageOutputBuilder はメッセージエンティティからMessageOutputを構築する補助コンポーネントです
type MessageOutputBuilder struct {
	messageRepo      domainrepository.MessageRepository
	userRepo         domainrepository.UserRepository
	userGroupRepo    domainrepository.UserGroupRepository
	userMentionRepo  domainrepository.MessageUserMentionRepository
	groupMentionRepo domainrepository.MessageGroupMentionRepository
	linkRepo         domainrepository.MessageLinkRepository
	attachmentRepo   domainrepository.AttachmentRepository
	assembler        *MessageOutputAssembler
}

// NewMessageOutputBuilder はMessageOutputBuilderを作成します
func NewMessageOutputBuilder(
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	assembler *MessageOutputAssembler,
) *MessageOutputBuilder {
	return &MessageOutputBuilder{
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		userGroupRepo:    userGroupRepo,
		userMentionRepo:  userMentionRepo,
		groupMentionRepo: groupMentionRepo,
		linkRepo:         linkRepo,
		attachmentRepo:   attachmentRepo,
		assembler:        assembler,
	}
}

// Build はメッセージ配列からMessageOutputスライスを構築します
func (b *MessageOutputBuilder) Build(ctx context.Context, messages []*entity.Message) ([]MessageOutput, error) {
	if len(messages) == 0 {
		return []MessageOutput{}, nil
	}

	messageIDs := make([]string, len(messages))
	for idx, msg := range messages {
		messageIDs[idx] = msg.ID
	}

	relatedData, err := b.fetchRelatedData(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	userMap, err := b.fetchUserMap(ctx, messages, relatedData.Reactions)
	if err != nil {
		return nil, err
	}

	groups, err := b.fetchGroups(ctx, relatedData.GroupMentions)
	if err != nil {
		return nil, err
	}

	userMentionsByMessage := b.groupUserMentionsByMessage(relatedData.UserMentions)
	groupMentionsByMessage := b.groupGroupMentionsByMessage(relatedData.GroupMentions)
	linksByMessage := b.groupLinksByMessage(relatedData.Links)

	outputs := make([]MessageOutput, 0, len(messages))
	for _, msg := range messages {
		outputs = append(outputs, b.assembler.AssembleMessageOutput(
			msg,
			userMap[msg.UserID],
			userMentionsByMessage[msg.ID],
			groupMentionsByMessage[msg.ID],
			linksByMessage[msg.ID],
			relatedData.Reactions[msg.ID],
			relatedData.Attachments[msg.ID],
			groups,
			userMap,
		))
	}

	return outputs, nil
}

// fetchRelatedData はメッセージに関連するデータを一括取得します
func (b *MessageOutputBuilder) fetchRelatedData(ctx context.Context, messageIDs []string) (*RelatedData, error) {
	if len(messageIDs) == 0 {
		return &RelatedData{
			UserMentions:  []*entity.MessageUserMention{},
			GroupMentions: []*entity.MessageGroupMention{},
			Links:         []*entity.MessageLink{},
			Reactions:     map[string][]*entity.MessageReaction{},
			Attachments:   map[string][]*entity.Attachment{},
		}, nil
	}

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

	return &RelatedData{
		UserMentions:  userMentions,
		GroupMentions: groupMentions,
		Links:         links,
		Reactions:     reactions,
		Attachments:   attachments,
	}, nil
}

// fetchUserMap はメッセージおよびリアクションに関係するユーザー情報を取得します
func (b *MessageOutputBuilder) fetchUserMap(ctx context.Context, messages []*entity.Message, reactions map[string][]*entity.MessageReaction) (map[string]*entity.User, error) {
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)

	for _, msg := range messages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
		if msg.DeletedBy != nil && !userIDSet[*msg.DeletedBy] {
			userIDs = append(userIDs, *msg.DeletedBy)
			userIDSet[*msg.DeletedBy] = true
		}
	}

	for _, reactionList := range reactions {
		for _, reaction := range reactionList {
			if !userIDSet[reaction.UserID] {
				userIDs = append(userIDs, reaction.UserID)
				userIDSet[reaction.UserID] = true
			}
		}
	}

	if len(userIDs) == 0 {
		return map[string]*entity.User{}, nil
	}

	users, err := b.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	return userMap, nil
}

// fetchGroups はグループメンションに関係するグループ情報を取得します
func (b *MessageOutputBuilder) fetchGroups(ctx context.Context, groupMentions []*entity.MessageGroupMention) (map[string]*entity.UserGroup, error) {
	groupIDs := make([]string, 0)
	groupIDSet := make(map[string]bool)

	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDs = append(groupIDs, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	if len(groupIDs) == 0 {
		return make(map[string]*entity.UserGroup), nil
	}

	groups, err := b.userGroupRepo.FindByIDs(ctx, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %w", err)
	}

	groupMap := make(map[string]*entity.UserGroup)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	return groupMap, nil
}

func (b *MessageOutputBuilder) groupUserMentionsByMessage(userMentions []*entity.MessageUserMention) map[string][]*entity.MessageUserMention {
	grouped := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		grouped[mention.MessageID] = append(grouped[mention.MessageID], mention)
	}
	return grouped
}

func (b *MessageOutputBuilder) groupGroupMentionsByMessage(groupMentions []*entity.MessageGroupMention) map[string][]*entity.MessageGroupMention {
	grouped := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		grouped[mention.MessageID] = append(grouped[mention.MessageID], mention)
	}
	return grouped
}

func (b *MessageOutputBuilder) groupLinksByMessage(links []*entity.MessageLink) map[string][]*entity.MessageLink {
	grouped := make(map[string][]*entity.MessageLink)
	for _, link := range links {
		grouped[link.MessageID] = append(grouped[link.MessageID], link)
	}
	return grouped
}
