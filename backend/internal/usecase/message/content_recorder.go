package message

import (
	"context"
	"fmt"
	"slices"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// ContentRecorder は本文のメンションとリンクを投稿時点の内容で保存します。投稿と編集で共有する
type ContentRecorder struct {
	mentionService        service.MentionService
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkProcessingService service.LinkProcessingService
	linkRepo              domainrepository.MessageLinkRepository
}

func NewContentRecorder(
	mentionService service.MentionService,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkProcessingService service.LinkProcessingService,
	linkRepo domainrepository.MessageLinkRepository,
) *ContentRecorder {
	return &ContentRecorder{
		mentionService:        mentionService,
		userMentionRepo:       userMentionRepo,
		groupMentionRepo:      groupMentionRepo,
		linkProcessingService: linkProcessingService,
		linkRepo:              linkRepo,
	}
}

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

// prepare はメンションを解決し、リンクのプレビューを取得します。OGP を取りに行くためトランザクションの外で呼びます
// previous は編集前のユーザーメンションで、本文に残ったグループは編集前に展開したメンバーを引き継ぎます
func (r *ContentRecorder) prepare(ctx context.Context, body, workspaceID string, previous []*entity.MessageUserMention) (*preparedContent, error) {
	var knownGroups []string
	for _, m := range previous {
		if m.ViaGroupID != nil && !slices.Contains(knownGroups, *m.ViaGroupID) {
			knownGroups = append(knownGroups, *m.ViaGroupID)
		}
	}
	resolved, err := r.mentionService.Resolve(ctx, body, workspaceID, knownGroups)
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

	if content.links, err = r.linkProcessingService.PrepareLinks(ctx, body, workspaceID); err != nil {
		return nil, fmt.Errorf("failed to prepare links: %w", err)
	}
	return content, nil
}

// save は prepare で用意したメンションとリンクをまとめて保存します。メッセージと同じトランザクションの中で呼びます
func (r *ContentRecorder) save(ctx context.Context, messageID string, content *preparedContent) error {
	for _, m := range content.userMentions {
		m.MessageID = messageID
	}
	for _, m := range content.groupMentions {
		m.MessageID = messageID
	}
	for _, l := range content.links {
		l.MessageID = messageID
	}
	if err := r.userMentionRepo.CreateBulk(ctx, content.userMentions); err != nil {
		return fmt.Errorf("failed to create user mentions: %w", err)
	}
	if err := r.groupMentionRepo.CreateBulk(ctx, content.groupMentions); err != nil {
		return fmt.Errorf("failed to create group mentions: %w", err)
	}
	if err := r.linkRepo.CreateBulk(ctx, content.links); err != nil {
		return fmt.Errorf("failed to create links: %w", err)
	}
	return nil
}

// replace は編集で本文が変わったメッセージのメンションとリンクを入れ替えます
func (r *ContentRecorder) replace(ctx context.Context, messageID string, content *preparedContent) error {
	if err := r.userMentionRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete user mentions: %w", err)
	}
	if err := r.groupMentionRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete group mentions: %w", err)
	}
	if err := r.linkRepo.DeleteByMessageID(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete links: %w", err)
	}
	return r.save(ctx, messageID, content)
}

// previousMentions は編集前のユーザーメンションです。グループ経由で展開したメンバーを引き継ぐのに使います
func (r *ContentRecorder) previousMentions(ctx context.Context, messageID string) ([]*entity.MessageUserMention, error) {
	mentions, err := r.userMentionRepo.FindByMessageIDs(ctx, []string{messageID})
	if err != nil {
		return nil, fmt.Errorf("failed to load user mentions: %w", err)
	}
	return mentions, nil
}
