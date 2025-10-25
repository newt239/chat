package mention

import (
	"context"
	"regexp"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type mentionService struct {
	workspaceRepo    repository.WorkspaceRepository
	userRepo         repository.UserRepository
	userGroupRepo    repository.UserGroupRepository
	userMentionRepo  repository.MessageUserMentionRepository
	groupMentionRepo repository.MessageGroupMentionRepository
}

func NewMentionService(
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
	userGroupRepo repository.UserGroupRepository,
	userMentionRepo repository.MessageUserMentionRepository,
	groupMentionRepo repository.MessageGroupMentionRepository,
) service.MentionService {
	return &mentionService{
		workspaceRepo:    workspaceRepo,
		userRepo:         userRepo,
		userGroupRepo:    userGroupRepo,
		userMentionRepo:  userMentionRepo,
		groupMentionRepo: groupMentionRepo,
	}
}

// ExtractUserMentions はメッセージ本文からユーザーメンションを抽出します
func (s *mentionService) ExtractUserMentions(ctx context.Context, body, workspaceID string) ([]*entity.MessageUserMention, error) {
	// @username パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*entity.MessageUserMention
	userIDSet := make(map[string]bool)

	// ワークスペースの全メンバーを一度に取得
	workspaceMembers, err := s.workspaceRepo.FindMembersByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return mentions, err
	}

	// メンバーのユーザーIDを収集
	userIDs := make([]string, 0, len(workspaceMembers))
	for _, member := range workspaceMembers {
		userIDs = append(userIDs, member.UserID)
	}

	// バルクでユーザー情報を取得
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return mentions, err
	}

	// ユーザーIDをキーとしたマップを作成
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンションを処理
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		username := match[1]

		// ユーザー名でマッチング
		for _, member := range workspaceMembers {
			user, exists := userMap[member.UserID]
			if !exists {
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

	return mentions, nil
}

// ExtractGroupMentions はメッセージ本文からグループメンションを抽出します
func (s *mentionService) ExtractGroupMentions(ctx context.Context, body, workspaceID string) ([]*entity.MessageGroupMention, error) {
	// @groupname パターンを検出
	mentionRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := mentionRegex.FindAllStringSubmatch(body, -1)

	var mentions []*entity.MessageGroupMention
	groupIDSet := make(map[string]bool)

	// ワークスペースの全グループを取得
	groups, err := s.userGroupRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return mentions, err
	}

	// グループ名をキーとしたマップを作成
	groupMap := make(map[string]*entity.UserGroup)
	for _, group := range groups {
		groupMap[group.Name] = group
	}

	// メンションを処理
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		groupname := match[1]

		// グループ名でマッチング
		if group, exists := groupMap[groupname]; exists {
			if !groupIDSet[group.ID] {
				mentions = append(mentions, &entity.MessageGroupMention{
					GroupID: group.ID,
				})
				groupIDSet[group.ID] = true
			}
		}
	}

	return mentions, nil
}
