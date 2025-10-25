package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"github.com/example/chat/internal/domain/service"
	"github.com/example/chat/internal/infrastructure/ogp"
)

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrParentMessageNotFound = errors.New("parent message not found")
	ErrMessageNotFound       = errors.New("message not found")
	ErrMessageAlreadyDeleted = errors.New("message already deleted")
	ErrCannotEditDeleted     = errors.New("cannot edit deleted message")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type MessageUseCase interface {
	ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error)
	CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error)
	UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error)
	DeleteMessage(ctx context.Context, input DeleteMessageInput) error
	GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error)
	GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error)
	ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error)
}

type messageInteractor struct {
	messageRepo      domainrepository.MessageRepository
	channelRepo      domainrepository.ChannelRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	userRepo         domainrepository.UserRepository
	userGroupRepo    domainrepository.UserGroupRepository
	userMentionRepo  domainrepository.MessageUserMentionRepository
	groupMentionRepo domainrepository.MessageGroupMentionRepository
	linkRepo         domainrepository.MessageLinkRepository
	threadRepo       domainrepository.ThreadRepository
	attachmentRepo   domainrepository.AttachmentRepository
	ogpService       *ogp.OGPService
	notificationSvc  service.NotificationService
}

func NewMessageInteractor(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	notificationSvc service.NotificationService,
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
		threadRepo:       threadRepo,
		attachmentRepo:   attachmentRepo,
		ogpService:       ogp.NewOGPService(),
		notificationSvc:  notificationSvc,
	}
}

func (i *messageInteractor) ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error) {
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
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

	messages, err := i.messageRepo.FindByChannelID(ctx, channel.ID, fetchLimit, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	// メッセージIDを収集
	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	// メンション、リンク、リアクションを一括取得
	userMentions, err := i.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user mentions: %w", err)
	}

	groupMentions, err := i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group mentions: %w", err)
	}

	links, err := i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}

	reactions, err := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}

	attachments, err := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attachments: %w", err)
	}

	// ユーザーIDを収集（メッセージ作成者とリアクションユーザー）
	userIDs := make([]string, 0, len(messages))
	userIDSet := make(map[string]bool)
	for _, msg := range messages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
	}
	// リアクションユーザーIDも追加
	for _, reactionList := range reactions {
		for _, reaction := range reactionList {
			if !userIDSet[reaction.UserID] {
				userIDs = append(userIDs, reaction.UserID)
				userIDSet[reaction.UserID] = true
			}
		}
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンション、リンク、リアクションをメッセージIDでグループ化
	userMentionsByMessage := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}

	groupMentionsByMessage := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}

	linksByMessage := make(map[string][]*entity.MessageLink)
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

	groups := make(map[string]*entity.UserGroup)
	if len(groupIDs) > 0 {
		// グループ情報を一括取得（簡略化のため、個別に取得）
		for _, groupID := range groupIDs {
			group, err := i.userGroupRepo.FindByID(ctx, groupID)
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
			reactions[msg.ID],
			attachments[msg.ID],
			groups,
			userMap,
		))
	}

	return &ListMessagesOutput{Messages: outputs, HasMore: hasMore}, nil
}

func (i *messageInteractor) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parent, err := i.messageRepo.FindByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != channel.ID {
			return nil, ErrParentMessageNotFound
		}
	}

	message := &entity.Message{
		ChannelID: channel.ID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Body:      input.Body,
		CreatedAt: time.Now(),
	}

	if err := i.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// 添付ファイルをメッセージに紐付け
	if len(input.AttachmentIDs) > 0 {
		if err := i.attachmentRepo.AttachToMessage(ctx, input.AttachmentIDs, message.ID); err != nil {
			return nil, fmt.Errorf("failed to attach files: %w", err)
		}
	}

	// スレッド返信の場合、メタデータを更新
	if input.ParentID != nil {
		if err := i.threadRepo.IncrementReplyCount(ctx, *input.ParentID, input.UserID); err != nil {
			// エラーが発生してもメッセージ作成は成功とする（ログ出力のみ）
			fmt.Printf("Warning: failed to update thread metadata: %v\n", err)
		}
	}

	// メンションとリンクを抽出・保存
	if err := i.extractAndSaveMentionsAndLinks(ctx, message.ID, input.Body, channel.WorkspaceID); err != nil {
		// エラーが発生してもメッセージ作成は成功とする（ログ出力のみ）
		fmt.Printf("Warning: failed to extract mentions and links: %v\n", err)
	}

	// ユーザー情報を取得
	user, err := i.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// メンションとリンクの情報を取得してレスポンスに含める
	userMentions, _ := i.userMentionRepo.FindByMessageID(ctx, message.ID)
	groupMentions, _ := i.groupMentionRepo.FindByMessageID(ctx, message.ID)
	links, _ := i.linkRepo.FindByMessageID(ctx, message.ID)
	attachmentList, _ := i.attachmentRepo.FindByMessageID(ctx, message.ID)

	// グループ情報を取得
	groups := make(map[string]*entity.UserGroup)
	for _, mention := range groupMentions {
		if group, err := i.userGroupRepo.FindByID(ctx, mention.GroupID); err == nil && group != nil {
			groups[mention.GroupID] = group
		}
	}

	// リアクションは新規作成メッセージには存在しないため空配列
	reactions := []*entity.MessageReaction{}

	// ユーザーマップを作成
	userMap := map[string]*entity.User{user.ID: user}

	output := toMessageOutputWithMentionsAndLinks(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)

	// WebSocket通知を送信（nilチェックを追加）
	if i.notificationSvc != nil {
		// outputをmap[string]interface{}に変換
		messageMap, err := convertStructToMap(output)
		if err == nil {
			i.notificationSvc.NotifyNewMessage(channel.WorkspaceID, channel.ID, messageMap)
		} else {
			fmt.Printf("Warning: failed to convert message to map: %v\n", err)
		}
	}

	return &output, nil
}

func (i *messageInteractor) ensureChannelAccess(ctx context.Context, channelID, userID string) (*entity.Channel, error) {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.IsPrivate {
		isMember, err := i.channelRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
		return ch, nil
	}

	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return ch, nil
}

func toMessageOutput(message *entity.Message, user *entity.User) MessageOutput {
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
		Reactions: []ReactionInfo{},
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		DeletedAt: message.DeletedAt,
	}
}

func toMessageOutputWithMentionsAndLinks(
	message *entity.Message,
	user *entity.User,
	userMentions []*entity.MessageUserMention,
	groupMentions []*entity.MessageGroupMention,
	links []*entity.MessageLink,
	reactions []*entity.MessageReaction,
	attachments []*entity.Attachment,
	groups map[string]*entity.UserGroup,
	userMap map[string]*entity.User,
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
	mentions := make([]UserMention, 0, len(userMentions))
	for _, mention := range userMentions {
		mentions = append(mentions, UserMention{
			UserID:      mention.UserID,
			DisplayName: "", // 必要に応じてユーザー情報を取得
		})
	}

	// グループメンションを変換
	groupMentionsOutput := make([]GroupMention, 0, len(groupMentions))
	for _, mention := range groupMentions {
		groupName := ""
		if group, exists := groups[mention.GroupID]; exists {
			groupName = group.Name
		}
		groupMentionsOutput = append(groupMentionsOutput, GroupMention{
			GroupID: mention.GroupID,
			Name:    groupName,
		})
	}

	// リンク情報を変換
	linksOutput := make([]LinkInfo, 0, len(links))
	for _, link := range links {
		linksOutput = append(linksOutput, LinkInfo{
			ID:          link.ID,
			URL:         link.URL,
			Title:       link.Title,
			Description: link.Description,
			ImageURL:    link.ImageURL,
			SiteName:    link.SiteName,
			CardType:    link.CardType,
		})
	}

	// リアクション情報を変換
	reactionsOutput := make([]ReactionInfo, 0, len(reactions))
	for _, reaction := range reactions {
		reactionUser, exists := userMap[reaction.UserID]
		reactionUserInfo := UserInfo{
			ID:          reaction.UserID,
			DisplayName: "Unknown User",
			AvatarURL:   nil,
		}
		if exists && reactionUser != nil {
			reactionUserInfo = UserInfo{
				ID:          reactionUser.ID,
				DisplayName: reactionUser.DisplayName,
				AvatarURL:   reactionUser.AvatarURL,
			}
		}
		reactionsOutput = append(reactionsOutput, ReactionInfo{
			User:      reactionUserInfo,
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}

	// 添付ファイル情報を変換
	attachmentsOutput := make([]AttachmentInfo, 0, len(attachments))
	for _, attachment := range attachments {
		attachmentsOutput = append(attachmentsOutput, AttachmentInfo{
			ID:        attachment.ID,
			FileName:  attachment.FileName,
			MimeType:  attachment.MimeType,
			SizeBytes: attachment.SizeBytes,
		})
	}

	return MessageOutput{
		ID:          message.ID,
		ChannelID:   message.ChannelID,
		UserID:      message.UserID,
		User:        userInfo,
		ParentID:    message.ParentID,
		Body:        message.Body,
		Mentions:    mentions,
		Groups:      groupMentionsOutput,
		Links:       linksOutput,
		Reactions:   reactionsOutput,
		Attachments: attachmentsOutput,
		CreatedAt:   message.CreatedAt,
		EditedAt:    message.EditedAt,
		DeletedAt:   message.DeletedAt,
	}
}

// メンションとリンクの抽出・保存
func (i *messageInteractor) extractAndSaveMentionsAndLinks(ctx context.Context, messageID, body, workspaceID string) error {
	// ユーザーメンションの抽出
	userMentions := i.extractUserMentions(ctx, body, workspaceID)
	for _, mention := range userMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.userMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	// グループメンションの抽出
	groupMentions := i.extractGroupMentions(ctx, body, workspaceID)
	for _, mention := range groupMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.groupMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	// リンクの抽出とOGP取得
	urls := ogp.ExtractURLs(body)
	for _, urlStr := range urls {
		// 既存のリンクをチェック
		existingLink, err := i.linkRepo.FindByURL(ctx, urlStr)
		if err != nil {
			continue // エラーは無視
		}

		var link *entity.MessageLink
		if existingLink != nil {
			// 既存のリンクを再利用
			link = &entity.MessageLink{
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

			link = &entity.MessageLink{
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

		if err := i.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}

// ユーザーメンションの抽出
func (i *messageInteractor) extractUserMentions(ctx context.Context, body, workspaceID string) []*entity.MessageUserMention {
	// @username パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*entity.MessageUserMention
	userIDSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		username := match[1]

		// ユーザー名でユーザーを検索（簡略化のため、display_nameで検索）
		// 実際の実装では、ユーザー名フィールドを追加するか、別の方法で検索
		// ここでは簡略化のため、ワークスペースの全ユーザーを取得して検索
		workspaceMembers, err := i.workspaceRepo.FindMembersByWorkspaceID(ctx, workspaceID)
		if err != nil {
			continue
		}

		for _, member := range workspaceMembers {
			user, err := i.userRepo.FindByID(ctx, member.UserID)
			if err != nil || user == nil {
				continue
			}
			// 簡略化のため、display_nameの最初の部分でマッチング
			if strings.HasPrefix(strings.ToLower(user.DisplayName), strings.ToLower(username)) {
				if !userIDSet[user.ID] {
					mentions = append(mentions, &entity.MessageUserMention{
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
func (i *messageInteractor) extractGroupMentions(ctx context.Context, body, workspaceID string) []*entity.MessageGroupMention {
	// @groupname パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*entity.MessageGroupMention
	groupIDSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		groupname := match[1]

		// グループ名でグループを検索
		group, err := i.userGroupRepo.FindByName(ctx, workspaceID, groupname)
		if err != nil || group == nil {
			continue
		}

		if !groupIDSet[group.ID] {
			mentions = append(mentions, &entity.MessageGroupMention{
				GroupID: group.ID,
			})
			groupIDSet[group.ID] = true
		}
	}

	return mentions
}

func (i *messageInteractor) GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error) {
	// 親メッセージを取得
	parentMessage, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parent message: %w", err)
	}
	if parentMessage == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
	_, err = i.ensureChannelAccess(ctx, parentMessage.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッド返信を取得
	replies, err := i.messageRepo.FindThreadReplies(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	// 親メッセージとリプライのメッセージIDを収集
	allMessages := append([]*entity.Message{parentMessage}, replies...)
	messageIDs := make([]string, len(allMessages))
	for idx, msg := range allMessages {
		messageIDs[idx] = msg.ID
	}

	// メンション、リンク、リアクション、添付ファイルを一括取得
	userMentions, _ := i.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	groupMentions, _ := i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	links, _ := i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	reactions, _ := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	attachments, _ := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs)

	// ユーザーIDを収集
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for _, msg := range allMessages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
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

	// ユーザー情報を一括取得
	users, _ := i.userRepo.FindByIDs(ctx, userIDs)
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンション、リンク、リアクションをメッセージIDでグループ化
	userMentionsByMessage := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}

	groupMentionsByMessage := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}

	linksByMessage := make(map[string][]*entity.MessageLink)
	for _, link := range links {
		linksByMessage[link.MessageID] = append(linksByMessage[link.MessageID], link)
	}

	// グループ情報を取得
	groupIDs := make([]string, 0)
	groupIDSet := make(map[string]bool)
	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDs = append(groupIDs, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	groups := make(map[string]*entity.UserGroup)
	for _, groupID := range groupIDs {
		group, err := i.userGroupRepo.FindByID(ctx, groupID)
		if err == nil && group != nil {
			groups[groupID] = group
		}
	}

	// 親メッセージ出力を作成
	parentOutput := toMessageOutputWithMentionsAndLinks(
		parentMessage,
		userMap[parentMessage.UserID],
		userMentionsByMessage[parentMessage.ID],
		groupMentionsByMessage[parentMessage.ID],
		linksByMessage[parentMessage.ID],
		reactions[parentMessage.ID],
		attachments[parentMessage.ID],
		groups,
		userMap,
	)

	// リプライ出力を作成
	replyOutputs := make([]MessageOutput, 0, len(replies))
	for _, reply := range replies {
		replyOutputs = append(replyOutputs, toMessageOutputWithMentionsAndLinks(
			reply,
			userMap[reply.UserID],
			userMentionsByMessage[reply.ID],
			groupMentionsByMessage[reply.ID],
			linksByMessage[reply.ID],
			reactions[reply.ID],
			attachments[reply.ID],
			groups,
			userMap,
		))
	}

	return &GetThreadRepliesOutput{
		ParentMessage: parentOutput,
		Replies:       replyOutputs,
		HasMore:       false,
	}, nil
}

func (i *messageInteractor) GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error) {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
	_, err = i.ensureChannelAccess(ctx, message.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッドメタデータを取得
	metadata, err := i.threadRepo.FindMetadataByMessageID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread metadata: %w", err)
	}

	// メタデータが存在しない場合は空のメタデータを返す
	if metadata == nil {
		return &ThreadMetadataOutput{
			MessageID:          input.MessageID,
			ReplyCount:         0,
			LastReplyAt:        nil,
			LastReplyUser:      nil,
			ParticipantUserIDs: []string{},
		}, nil
	}

	// 最新返信者の情報を取得
	var lastReplyUser *UserInfo
	if metadata.LastReplyUserID != nil {
		user, err := i.userRepo.FindByID(ctx, *metadata.LastReplyUserID)
		if err == nil && user != nil {
			lastReplyUser = &UserInfo{
				ID:          user.ID,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			}
		}
	}

	return &ThreadMetadataOutput{
		MessageID:          metadata.MessageID,
		ReplyCount:         metadata.ReplyCount,
		LastReplyAt:        metadata.LastReplyAt,
		LastReplyUser:      lastReplyUser,
		ParticipantUserIDs: metadata.ParticipantUserIDs,
	}, nil
}

// convertStructToMap は構造体をmap[string]interface{}に変換します
func convertStructToMap(data interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (i *messageInteractor) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error) {
	// 通常のメッセージ一覧を取得
	listOutput, err := i.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}

	// メッセージIDを収集
	messageIDs := make([]string, len(listOutput.Messages))
	for idx, msg := range listOutput.Messages {
		messageIDs[idx] = msg.ID
	}

	// スレッドメタデータを一括取得
	metadataMap, err := i.threadRepo.FindMetadataByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread metadata: %w", err)
	}

	// 最新返信者のユーザーIDを収集
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for _, metadata := range metadataMap {
		if metadata.LastReplyUserID != nil && !userIDSet[*metadata.LastReplyUserID] {
			userIDs = append(userIDs, *metadata.LastReplyUserID)
			userIDSet[*metadata.LastReplyUserID] = true
		}
	}

	// ユーザー情報を一括取得
	users, _ := i.userRepo.FindByIDs(ctx, userIDs)
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メッセージとスレッドメタデータを結合
	outputs := make([]MessageWithThreadOutput, 0, len(listOutput.Messages))
	for _, msg := range listOutput.Messages {
		output := MessageWithThreadOutput{
			MessageOutput: msg,
		}

		if metadata, exists := metadataMap[msg.ID]; exists {
			var lastReplyUser *UserInfo
			if metadata.LastReplyUserID != nil {
				user := userMap[*metadata.LastReplyUserID]
				if user != nil {
					lastReplyUser = &UserInfo{
						ID:          user.ID,
						DisplayName: user.DisplayName,
						AvatarURL:   user.AvatarURL,
					}
				}
			}

			output.ThreadMetadata = &ThreadMetadataOutput{
				MessageID:          metadata.MessageID,
				ReplyCount:         metadata.ReplyCount,
				LastReplyAt:        metadata.LastReplyAt,
				LastReplyUser:      lastReplyUser,
				ParticipantUserIDs: metadata.ParticipantUserIDs,
			}
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (i *messageInteractor) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	// メッセージ存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// チャンネルアクセス確認
	channel, err := i.ensureChannelAccess(ctx, message.ChannelID, input.EditorID)
	if err != nil {
		return nil, err
	}

	// 削除済みメッセージの編集禁止
	if message.DeletedAt != nil {
		return nil, ErrCannotEditDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canEdit, err := i.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.EditorID)
	if err != nil {
		return nil, fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canEdit {
		return nil, ErrUnauthorized
	}

	// メッセージ本文を更新
	message.Body = input.Body
	now := time.Now()
	message.EditedAt = &now

	// データベース更新
	if err := i.messageRepo.Update(ctx, message); err != nil {
		return nil, fmt.Errorf("メッセージの更新に失敗しました: %w", err)
	}

	// 既存のメンション・リンクを削除
	if err := i.userMentionRepo.DeleteByMessageID(ctx, message.ID); err != nil {
		fmt.Printf("Warning: failed to delete user mentions: %v\n", err)
	}
	if err := i.groupMentionRepo.DeleteByMessageID(ctx, message.ID); err != nil {
		fmt.Printf("Warning: failed to delete group mentions: %v\n", err)
	}
	if err := i.linkRepo.DeleteByMessageID(ctx, message.ID); err != nil {
		fmt.Printf("Warning: failed to delete links: %v\n", err)
	}

	// 新しいメンション・リンクを抽出・保存
	if err := i.extractAndSaveMentionsAndLinks(ctx, message.ID, input.Body, channel.WorkspaceID); err != nil {
		fmt.Printf("Warning: failed to extract and save mentions/links: %v\n", err)
	}

	// 更新後のデータを取得してMessageOutputを構築
	userMentions, err := i.userMentionRepo.FindByMessageID(ctx, message.ID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch user mentions: %v\n", err)
		userMentions = []*entity.MessageUserMention{}
	}

	groupMentions, err := i.groupMentionRepo.FindByMessageID(ctx, message.ID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch group mentions: %v\n", err)
		groupMentions = []*entity.MessageGroupMention{}
	}

	links, err := i.linkRepo.FindByMessageID(ctx, message.ID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch links: %v\n", err)
		links = []*entity.MessageLink{}
	}

	reactions, err := i.messageRepo.FindReactions(ctx, message.ID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch reactions: %v\n", err)
		reactions = []*entity.MessageReaction{}
	}

	attachmentList, err := i.attachmentRepo.FindByMessageID(ctx, message.ID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch attachments: %v\n", err)
		attachmentList = []*entity.Attachment{}
	}

	// ユーザー情報を取得
	user, err := i.userRepo.FindByID(ctx, message.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	// グループ情報を取得
	groups := make(map[string]*entity.UserGroup)
	for _, gm := range groupMentions {
		group, err := i.userGroupRepo.FindByID(ctx, gm.GroupID)
		if err == nil && group != nil {
			groups[gm.GroupID] = group
		}
	}

	userMap := map[string]*entity.User{user.ID: user}
	output := toMessageOutputWithMentionsAndLinks(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)

	// WebSocket通知を送信
	if i.notificationSvc != nil {
		messageMap, err := convertStructToMap(output)
		if err == nil {
			i.notificationSvc.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, messageMap)
		} else {
			fmt.Printf("Warning: failed to convert message to map: %v\n", err)
		}
	}

	return &output, nil
}

func (i *messageInteractor) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	// メッセージ存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルアクセス確認
	channel, err := i.ensureChannelAccess(ctx, message.ChannelID, input.ExecutorID)
	if err != nil {
		return err
	}

	// 既に削除済みの場合はエラー
	if message.DeletedAt != nil {
		return ErrMessageAlreadyDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canDelete, err := i.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.ExecutorID)
	if err != nil {
		return fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canDelete {
		return ErrUnauthorized
	}

	// 削除対象メッセージIDのリストを作成
	deleteIDs := []string{message.ID}

	// スレッド親メッセージの場合、子メッセージも削除
	if message.ParentID == nil {
		replies, err := i.messageRepo.FindThreadReplies(ctx, message.ID)
		if err != nil {
			return fmt.Errorf("返信の取得に失敗しました: %w", err)
		}
		for _, reply := range replies {
			deleteIDs = append(deleteIDs, reply.ID)
		}

		// スレッドメタデータも削除
		if err := i.threadRepo.DeleteMetadata(ctx, message.ID); err != nil {
			fmt.Printf("Warning: failed to delete thread metadata: %v\n", err)
		}
	}

	// ソフトデリート実行
	if err := i.messageRepo.SoftDeleteByIDs(ctx, deleteIDs, input.ExecutorID); err != nil {
		return fmt.Errorf("メッセージの削除に失敗しました: %w", err)
	}

	// WebSocket通知を送信
	if i.notificationSvc != nil {
		deleteData := map[string]interface{}{
			"messageId":  message.ID,
			"channelId":  message.ChannelID,
			"deletedIds": deleteIDs,
		}
		i.notificationSvc.NotifyDeletedMessage(channel.WorkspaceID, channel.ID, deleteData)
	}

	return nil
}

// canModifyMessage はユーザーがメッセージを編集・削除できるかどうかを確認します
func (i *messageInteractor) canModifyMessage(ctx context.Context, workspaceID, messageOwnerID, executorID string) (bool, error) {
	// 投稿者本人の場合は許可
	if messageOwnerID == executorID {
		return true, nil
	}

	// 管理者権限チェック
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, executorID)
	if err != nil {
		return false, fmt.Errorf("ワークスペースメンバー情報の取得に失敗しました: %w", err)
	}
	if member == nil {
		return false, nil
	}

	// owner または admin の場合は許可
	if member.Role == entity.WorkspaceRoleOwner || member.Role == entity.WorkspaceRoleAdmin {
		return true, nil
	}

	return false, nil
}
