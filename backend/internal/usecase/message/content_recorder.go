package message

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// contentRecorder は本文のメンションとリンクを投稿時点の内容で保存します
type contentRecorder struct {
	mentionService        service.MentionService
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkProcessingService service.LinkProcessingService
	linkRepo              domainrepository.MessageLinkRepository
}

// applyBroadcastMentions は本文の <@channel> / <@here> をメッセージに反映します。保存する前に呼びます
func applyBroadcastMentions(message *entity.Message) {
	tokens := entity.ParseMentionTokens(message.Body)
	message.MentionsChannel, message.MentionsHere = tokens.Channel, tokens.Here
}

// record はメンションとリンクを保存します。previous は編集前のユーザーメンションで、本文に残ったグループは編集前に展開したメンバーを引き継ぎます
func (r *contentRecorder) record(ctx context.Context, messageID, body, workspaceID string, previous []*entity.MessageUserMention) error {
	var knownGroups []string
	for _, m := range previous {
		if m.ViaGroupID != nil && !slices.Contains(knownGroups, *m.ViaGroupID) {
			knownGroups = append(knownGroups, *m.ViaGroupID)
		}
	}
	resolved, err := r.mentionService.Resolve(ctx, body, workspaceID, knownGroups)
	if err != nil {
		return fmt.Errorf("failed to resolve mentions: %w", err)
	}

	now := time.Now()
	var userMentions []*entity.MessageUserMention
	// 本人への直接のメンションをグループ経由より優先して 1 行にする
	add := func(userID string, viaGroupID *string) {
		for _, m := range userMentions {
			if m.UserID == userID {
				return
			}
		}
		userMentions = append(userMentions, &entity.MessageUserMention{MessageID: messageID, UserID: userID, ViaGroupID: viaGroupID, CreatedAt: now})
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
	}
	for _, mention := range userMentions {
		if err := r.userMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}
	for _, groupID := range resolved.GroupIDs {
		if err := r.groupMentionRepo.Create(ctx, &entity.MessageGroupMention{MessageID: messageID, GroupID: groupID, CreatedAt: now}); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	links, err := r.linkProcessingService.ProcessLinks(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to process links: %w", err)
	}
	for _, link := range links {
		link.MessageID = messageID
		if err := r.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}
	return nil
}
