package message

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/ogp"
)

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrParentMessageNotFound = errors.New("parent message not found")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type MessageUseCase interface {
	ListMessages(input ListMessagesInput) (*ListMessagesOutput, error)
	CreateMessage(input CreateMessageInput) (*MessageOutput, error)
}

type messageInteractor struct {
	messageRepo           domain.MessageRepository
	channelRepo           domain.ChannelRepository
	workspaceRepo         domain.WorkspaceRepository
	userRepo              domain.UserRepository
	userGroupRepo         domain.UserGroupRepository
	userMentionRepo       domain.MessageUserMentionRepository
	groupMentionRepo      domain.MessageGroupMentionRepository
	linkRepo              domain.MessageLinkRepository
	ogpService            *ogp.OGPService
}

func NewMessageInteractor(
	messageRepo domain.MessageRepository,
	channelRepo domain.ChannelRepository,
	workspaceRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
	userGroupRepo domain.UserGroupRepository,
	userMentionRepo domain.MessageUserMentionRepository,
	groupMentionRepo domain.MessageGroupMentionRepository,
	linkRepo domain.MessageLinkRepository,
) MessageUseCase {
	return &messageInteractor{
		messageRepo:      messageRepo,
		channelRepo:      channelRepo,
		workspaceRepo:    workspaceRepo,
		userRepo:         userRepo,
		userGroupRepo:    userGroupRepo,
		userMentionRepo:  userMentionRepo,
		groupMentionRepo: groupMentionRepo,
		linkRepo:         linkRepo,
		ogpService:       ogp.NewOGPService(),
	}
}

func (i *messageInteractor) ListMessages(input ListMessagesInput) (*ListMessagesOutput, error) {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	fetchLimit := limit + 1

	messages, err := i.messageRepo.FindByChannelID(channel.ID, fetchLimit, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	// ユーザーIDを収集
	userIDs := make([]string, 0, len(messages))
	userIDSet := make(map[string]bool)
	for _, msg := range messages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メッセージIDを収集
	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	// メンションとリンクを一括取得
	userMentions, err := i.userMentionRepo.FindByMessageIDs(messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user mentions: %w", err)
	}

	groupMentions, err := i.groupMentionRepo.FindByMessageIDs(messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group mentions: %w", err)
	}

	links, err := i.linkRepo.FindByMessageIDs(messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}

	// メンションとリンクをメッセージIDでグループ化
	userMentionsByMessage := make(map[string][]*domain.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}

	groupMentionsByMessage := make(map[string][]*domain.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}

	linksByMessage := make(map[string][]*domain.MessageLink)
	for _, link := range links {
		linksByMessage[link.MessageID] = append(linksByMessage[link.MessageID], link)
	}

	// グループ情報を取得（グループメンションがある場合）
	groupIDs := make([]string, 0)
	groupIDSet := make(map[string]bool)
	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDs = append(groupIDs, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	groups := make(map[string]*domain.UserGroup)
	if len(groupIDs) > 0 {
		// グループ情報を一括取得（簡略化のため、個別に取得）
		for _, groupID := range groupIDs {
			group, err := i.userGroupRepo.FindByID(groupID)
			if err == nil && group != nil {
				groups[groupID] = group
			}
		}
	}

	outputs := make([]MessageOutput, 0, len(messages))
	for _, msg := range messages {
		user := userMap[msg.UserID]
		outputs = append(outputs, toMessageOutputWithMentionsAndLinks(
			msg, 
			user, 
			userMentionsByMessage[msg.ID], 
			groupMentionsByMessage[msg.ID], 
			linksByMessage[msg.ID],
			groups,
		))
	}

	return &ListMessagesOutput{Messages: outputs, HasMore: hasMore}, nil
}

func (i *messageInteractor) CreateMessage(input CreateMessageInput) (*MessageOutput, error) {
	channel, err := i.ensureChannelAccess(input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parent, err := i.messageRepo.FindByID(*input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != channel.ID {
			return nil, ErrParentMessageNotFound
		}
	}

	message := &domain.Message{
		ChannelID: channel.ID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Body:      input.Body,
		CreatedAt: time.Now(),
	}

	if err := i.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// メンションとリンクを抽出・保存
	ctx := context.Background()
	if err := i.extractAndSaveMentionsAndLinks(ctx, message.ID, input.Body, channel.WorkspaceID); err != nil {
		// エラーが発生してもメッセージ作成は成功とする（ログ出力のみ）
		fmt.Printf("Warning: failed to extract mentions and links: %v\n", err)
	}

	// ユーザー情報を取得
	user, err := i.userRepo.FindByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// メンションとリンクの情報を取得してレスポンスに含める
	userMentions, _ := i.userMentionRepo.FindByMessageID(message.ID)
	groupMentions, _ := i.groupMentionRepo.FindByMessageID(message.ID)
	links, _ := i.linkRepo.FindByMessageID(message.ID)

	// グループ情報を取得
	groups := make(map[string]*domain.UserGroup)
	for _, mention := range groupMentions {
		if group, err := i.userGroupRepo.FindByID(mention.GroupID); err == nil && group != nil {
			groups[mention.GroupID] = group
		}
	}

	output := toMessageOutputWithMentionsAndLinks(message, user, userMentions, groupMentions, links, groups)
	return &output, nil
}

func (i *messageInteractor) ensureChannelAccess(channelID, userID string) (*domain.Channel, error) {
	ch, err := i.channelRepo.FindByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.IsPrivate {
		isMember, err := i.channelRepo.IsMember(ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
		return ch, nil
	}

	member, err := i.workspaceRepo.FindMember(ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return ch, nil
}

func toMessageOutput(message *domain.Message, user *domain.User) MessageOutput {
	userInfo := UserInfo{
		ID:          "",
		DisplayName: "Unknown User",
		AvatarURL:   nil,
	}

	if user != nil {
		userInfo = UserInfo{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	return MessageOutput{
		ID:        message.ID,
		ChannelID: message.ChannelID,
		UserID:    message.UserID,
		User:      userInfo,
		ParentID:  message.ParentID,
		Body:      message.Body,
		Mentions:  []UserMention{},
		Groups:    []GroupMention{},
		Links:     []LinkInfo{},
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		DeletedAt: message.DeletedAt,
	}
}

func toMessageOutputWithMentionsAndLinks(
	message *domain.Message,
	user *domain.User,
	userMentions []*domain.MessageUserMention,
	groupMentions []*domain.MessageGroupMention,
	links []*domain.MessageLink,
	groups map[string]*domain.UserGroup,
) MessageOutput {
	userInfo := UserInfo{
		ID:          "",
		DisplayName: "Unknown User",
		AvatarURL:   nil,
	}

	if user != nil {
		userInfo = UserInfo{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	// ユーザーメンションを変換
	mentions := make([]UserMention, len(userMentions))
	for i, mention := range userMentions {
		mentions[i] = UserMention{
			UserID:      mention.UserID,
			DisplayName: "", // 必要に応じてユーザー情報を取得
		}
	}

	// グループメンションを変換
	groupMentionsOutput := make([]GroupMention, len(groupMentions))
	for i, mention := range groupMentions {
		groupName := ""
		if group, exists := groups[mention.GroupID]; exists {
			groupName = group.Name
		}
		groupMentionsOutput[i] = GroupMention{
			GroupID: mention.GroupID,
			Name:    groupName,
		}
	}

	// リンク情報を変換
	linksOutput := make([]LinkInfo, len(links))
	for i, link := range links {
		linksOutput[i] = LinkInfo{
			ID:          link.ID,
			URL:         link.URL,
			Title:       link.Title,
			Description: link.Description,
			ImageURL:    link.ImageURL,
			SiteName:    link.SiteName,
			CardType:    link.CardType,
		}
	}

	return MessageOutput{
		ID:        message.ID,
		ChannelID: message.ChannelID,
		UserID:    message.UserID,
		User:      userInfo,
		ParentID:  message.ParentID,
		Body:      message.Body,
		Mentions:  mentions,
		Groups:    groupMentionsOutput,
		Links:     linksOutput,
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		DeletedAt: message.DeletedAt,
	}
}

// メンションとリンクの抽出・保存
func (i *messageInteractor) extractAndSaveMentionsAndLinks(ctx context.Context, messageID, body, workspaceID string) error {
	// ユーザーメンションの抽出
	userMentions := i.extractUserMentions(body, workspaceID)
	for _, mention := range userMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.userMentionRepo.Create(mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	// グループメンションの抽出
	groupMentions := i.extractGroupMentions(body, workspaceID)
	for _, mention := range groupMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.groupMentionRepo.Create(mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	// リンクの抽出とOGP取得
	urls := ogp.ExtractURLs(body)
	for _, urlStr := range urls {
		// 既存のリンクをチェック
		existingLink, err := i.linkRepo.FindByURL(urlStr)
		if err != nil {
			continue // エラーは無視
		}

		var link *domain.MessageLink
		if existingLink != nil {
			// 既存のリンクを再利用
			link = &domain.MessageLink{
				MessageID:   messageID,
				URL:         existingLink.URL,
				Title:       existingLink.Title,
				Description: existingLink.Description,
				ImageURL:    existingLink.ImageURL,
				SiteName:    existingLink.SiteName,
				CardType:    existingLink.CardType,
				CreatedAt:   time.Now(),
			}
		} else {
			// 新しいリンクのOGPを取得
			ogpData, err := i.ogpService.FetchOGP(ctx, urlStr)
			if err != nil {
				// OGP取得に失敗してもリンクは保存
				ogpData = &ogp.OGPData{}
			}

			link = &domain.MessageLink{
				MessageID:   messageID,
				URL:         urlStr,
				Title:       ogpData.Title,
				Description: ogpData.Description,
				ImageURL:    ogpData.ImageURL,
				SiteName:    ogpData.SiteName,
				CardType:    ogpData.CardType,
				CreatedAt:   time.Now(),
			}
		}

		if err := i.linkRepo.Create(link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}

// ユーザーメンションの抽出
func (i *messageInteractor) extractUserMentions(body, workspaceID string) []*domain.MessageUserMention {
	// @username パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*domain.MessageUserMention
	userIDSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		username := match[1]

		// ユーザー名でユーザーを検索（簡略化のため、display_nameで検索）
		// 実際の実装では、ユーザー名フィールドを追加するか、別の方法で検索
		// ここでは簡略化のため、ワークスペースの全ユーザーを取得して検索
		workspaceMembers, err := i.workspaceRepo.FindMembersByWorkspaceID(workspaceID)
		if err != nil {
			continue
		}

		for _, member := range workspaceMembers {
			user, err := i.userRepo.FindByID(member.UserID)
			if err != nil || user == nil {
				continue
			}
			// 簡略化のため、display_nameの最初の部分でマッチング
			if strings.HasPrefix(strings.ToLower(user.DisplayName), strings.ToLower(username)) {
				if !userIDSet[user.ID] {
					mentions = append(mentions, &domain.MessageUserMention{
						UserID: user.ID,
					})
					userIDSet[user.ID] = true
				}
				break
			}
		}
	}

	return mentions
}

// グループメンションの抽出
func (i *messageInteractor) extractGroupMentions(body, workspaceID string) []*domain.MessageGroupMention {
	// @groupname パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*domain.MessageGroupMention
	groupIDSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		groupname := match[1]

		// グループ名でグループを検索
		group, err := i.userGroupRepo.FindByName(workspaceID, groupname)
		if err != nil || group == nil {
			continue
		}

		if !groupIDSet[group.ID] {
			mentions = append(mentions, &domain.MessageGroupMention{
				GroupID: group.ID,
			})
			groupIDSet[group.ID] = true
		}
	}

	return mentions
}
