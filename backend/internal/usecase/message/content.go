package message

import (
	"context"
	"fmt"
	"slices"

	"github.com/newt239/chat/internal/domain/entity"
)

// preparedContent は保存する前に解決したメンションとリンクです
type preparedContent struct {
	userMentions  []*entity.MessageUserMention
	groupMentions []*entity.MessageGroupMention
	links         []*entity.MessageLink
}

// applyBroadcastMentions は本文の <@channel> / <@here> をメッセージに反映します。保存する前に呼びます
func applyBroadcastMentions(message *entity.Message) {
	tokens := entity.ParseMentionTokens(message.Body)
	message.MentionsChannel, message.MentionsHere = tokens.Channel, tokens.Here
}

// prepareContent は OGP を取りに行くためトランザクションの外で呼びます。本文に残ったグループは previous で展開済みのメンバーを引き継ぐ
func (i *Interactor) prepareContent(ctx context.Context, body, workspaceID string, previous []*entity.MessageUserMention) (*preparedContent, error) {
	var knownGroups []string
	for _, m := range previous {
		if m.ViaGroupID != nil && !slices.Contains(knownGroups, *m.ViaGroupID) {
			knownGroups = append(knownGroups, *m.ViaGroupID)
		}
	}
	resolved, err := i.mentionSvc.Resolve(ctx, body, workspaceID, knownGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mentions: %w", err)
	}

	content := &preparedContent{}
	// 本人への直接のメンションをグループ経由より優先して 1 行にする
	add := func(userID string, viaGroupID *string) {
		if !slices.ContainsFunc(content.userMentions, func(m *entity.MessageUserMention) bool { return m.UserID == userID }) {
			content.userMentions = append(content.userMentions, &entity.MessageUserMention{UserID: userID, ViaGroupID: viaGroupID})
		}
	}
	for _, userID := range resolved.UserIDs {
		add(userID, nil)
	}
	for _, m := range previous {
		if m.ViaGroupID != nil && slices.Contains(resolved.GroupIDs, *m.ViaGroupID) {
			add(m.UserID, m.ViaGroupID)
		}
	}
	for _, groupID := range resolved.GroupIDs {
		for _, userID := range resolved.GroupMembers[groupID] {
			add(userID, &groupID)
		}
		content.groupMentions = append(content.groupMentions, &entity.MessageGroupMention{GroupID: groupID})
	}

	if content.links, err = i.linkSvc.PrepareLinks(ctx, body, workspaceID); err != nil {
		return nil, fmt.Errorf("failed to prepare links: %w", err)
	}
	return content, nil
}

// saveContent はメッセージと同じトランザクションの中で呼びます
func (i *Interactor) saveContent(ctx context.Context, messageID string, content *preparedContent) error {
	for _, m := range content.userMentions {
		m.MessageID = messageID
	}
	for _, m := range content.groupMentions {
		m.MessageID = messageID
	}
	for _, l := range content.links {
		l.MessageID = messageID
	}
	if err := i.userMentionRepo.CreateBulk(ctx, content.userMentions); err != nil {
		return fmt.Errorf("failed to create user mentions: %w", err)
	}
	if err := i.groupMentionRepo.CreateBulk(ctx, content.groupMentions); err != nil {
		return fmt.Errorf("failed to create group mentions: %w", err)
	}
	if err := i.linkRepo.CreateBulk(ctx, content.links); err != nil {
		return fmt.Errorf("failed to create links: %w", err)
	}
	return nil
}

// replaceContent は編集で本文が変わったメッセージのメンションとリンクを入れ替えます
func (i *Interactor) replaceContent(ctx context.Context, messageID string, content *preparedContent) error {
	if err := i.userMentionRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete user mentions: %w", err)
	}
	if err := i.groupMentionRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete group mentions: %w", err)
	}
	if err := i.linkRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete links: %w", err)
	}
	return i.saveContent(ctx, messageID, content)
}
