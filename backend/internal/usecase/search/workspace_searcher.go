package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

const (
	defaultPerPage = 20
	maxPerPage     = 50
)

type WorkspaceSearcher struct {
	workspaceRepo        domainrepository.WorkspaceRepository
	channelRepo          domainrepository.ChannelRepository
	messageRepo          domainrepository.MessageRepository
	userRepo             domainrepository.UserRepository
	userGroupRepo        domainrepository.UserGroupRepository
	messageOutputBuilder *messageuc.MessageOutputBuilder
}

func NewWorkspaceSearcher(
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	messageOutputBuilder *messageuc.MessageOutputBuilder,
) *WorkspaceSearcher {
	return &WorkspaceSearcher{
		workspaceRepo:        workspaceRepo,
		channelRepo:          channelRepo,
		messageRepo:          messageRepo,
		userRepo:             userRepo,
		userGroupRepo:        userGroupRepo,
		messageOutputBuilder: messageOutputBuilder,
	}
}

func (s *WorkspaceSearcher) SearchWorkspace(ctx context.Context, input WorkspaceSearchInput) (*WorkspaceSearchOutput, error) {
	trimmedQuery := strings.TrimSpace(input.Query)
	if trimmedQuery == "" {
		return nil, ErrInvalidQuery
	}

	filter := input.Filter.Normalize()

	page := input.Page
	if page < 1 {
		page = 1
	}

	perPage := input.PerPage
	if perPage <= 0 {
		perPage = defaultPerPage
	} else if perPage > maxPerPage {
		perPage = maxPerPage
	}

	offset := (page - 1) * perPage

	workspace, err := s.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	member, err := s.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	messagesResult := PaginatedMessages{
		Items:   []messageuc.MessageOutput{},
		Total:   0,
		Page:    page,
		PerPage: perPage,
		HasMore: false,
	}

	channelsResult := PaginatedChannels{
		Items:   []channeluc.ChannelOutput{},
		Total:   0,
		Page:    page,
		PerPage: perPage,
		HasMore: false,
	}

	usersResult := PaginatedUsers{
		Items:   []workspaceuc.MemberInfo{},
		Total:   0,
		Page:    page,
		PerPage: perPage,
		HasMore: false,
	}

	groupsResult := PaginatedUserGroups{
		Items:   []usergroupuc.UserGroupOutput{},
		Total:   0,
		Page:    page,
		PerPage: perPage,
		HasMore: false,
	}

	// メッセージ検索
	if filter.includesMessages() {
		messagesResult, err = s.searchMessages(ctx, trimmedQuery, input.WorkspaceID, input.RequesterID, page, perPage, offset)
		if err != nil {
			return nil, err
		}
	}

	// チャンネル検索
	if filter.includesChannels() {
		channelsResult, err = s.searchChannels(ctx, trimmedQuery, input.WorkspaceID, input.RequesterID, page, perPage, offset)
		if err != nil {
			return nil, err
		}
	}

	// ユーザー検索
	if filter.includesUsers() {
		usersResult, err = s.searchUsers(ctx, trimmedQuery, input.WorkspaceID, page, perPage, offset)
		if err != nil {
			return nil, err
		}
	}

	// ユーザーグループ検索
	if filter.includesGroups() {
		groupsResult, err = s.searchUserGroups(ctx, trimmedQuery, input.WorkspaceID, page, perPage, offset)
		if err != nil {
			return nil, err
		}
	}

	return &WorkspaceSearchOutput{
		Messages: messagesResult,
		Channels: channelsResult,
		Users:    usersResult,
		Groups:   groupsResult,
	}, nil
}

func (s *WorkspaceSearcher) searchMessages(
	ctx context.Context,
	query string,
	workspaceID string,
	userID string,
	page int,
	limit int,
	offset int,
) (PaginatedMessages, error) {
	accessibleChannels, err := s.channelRepo.FindAccessibleChannels(ctx, workspaceID, userID)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to load accessible channels: %w", err)
	}

	channelIDs := make([]string, 0, len(accessibleChannels))
	for _, ch := range accessibleChannels {
		channelIDs = append(channelIDs, ch.ID)
	}

	result := PaginatedMessages{
		Items:   []messageuc.MessageOutput{},
		Total:   0,
		Page:    page,
		PerPage: limit,
		HasMore: false,
	}

	if len(channelIDs) == 0 {
		return result, nil
	}

	messages, total, err := s.messageRepo.SearchByChannelIDs(ctx, channelIDs, query, limit, offset)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to search messages: %w", err)
	}

	outputs, err := s.messageOutputBuilder.Build(ctx, userID, messages)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to build message outputs: %w", err)
	}

	result.Items = outputs
	result.Total = total
	result.HasMore = offset+len(outputs) < total

	return result, nil
}

func (s *WorkspaceSearcher) searchChannels(
	ctx context.Context,
	query string,
	workspaceID string,
	userID string,
	page int,
	limit int,
	offset int,
) (PaginatedChannels, error) {
	channels, total, err := s.channelRepo.SearchAccessibleChannels(ctx, workspaceID, userID, query, limit, offset)
	if err != nil {
		return PaginatedChannels{}, fmt.Errorf("failed to search channels: %w", err)
	}

	items := make([]channeluc.ChannelOutput, 0, len(channels))
	for _, ch := range channels {
		items = append(items, channeluc.ChannelOutput{
			ID:          ch.ID,
			WorkspaceID: ch.WorkspaceID,
			Name:        ch.Name,
			Description: ch.Description,
			IsPrivate:   ch.IsPrivate,
			CreatedBy:   ch.CreatedBy,
			CreatedAt:   ch.CreatedAt,
			UpdatedAt:   ch.UpdatedAt,
		})
	}

	return PaginatedChannels{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: limit,
		HasMore: offset+len(items) < total,
	}, nil
}

// searchUserGroups はワークスペース内のユーザーグループを名前で部分一致検索します
func (s *WorkspaceSearcher) searchUserGroups(
	ctx context.Context,
	query string,
	workspaceID string,
	page int,
	limit int,
	offset int,
) (PaginatedUserGroups, error) {
	groups, err := s.userGroupRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return PaginatedUserGroups{}, fmt.Errorf("failed to load user groups: %w", err)
	}

	lowered := strings.ToLower(query)
	matched := make([]usergroupuc.UserGroupOutput, 0, len(groups))
	for _, group := range groups {
		if !strings.Contains(strings.ToLower(group.Name), lowered) {
			continue
		}
		matched = append(matched, usergroupuc.UserGroupOutput{
			ID:          group.ID,
			WorkspaceID: group.WorkspaceID,
			Name:        group.Name,
			Description: group.Description,
			CreatedBy:   group.CreatedBy,
			CreatedAt:   group.CreatedAt,
			UpdatedAt:   group.UpdatedAt,
		})
	}

	total := len(matched)
	if offset > total {
		offset = total
	}
	end := min(offset+limit, total)

	return PaginatedUserGroups{
		Items:   matched[offset:end],
		Total:   total,
		Page:    page,
		PerPage: limit,
		HasMore: end < total,
	}, nil
}

func (s *WorkspaceSearcher) searchUsers(
	ctx context.Context,
	query string,
	workspaceID string,
	page int,
	limit int,
	offset int,
) (PaginatedUsers, error) {
	members, total, err := s.workspaceRepo.SearchMembers(ctx, workspaceID, query, limit, offset)
	if err != nil {
		return PaginatedUsers{}, fmt.Errorf("failed to search members: %w", err)
	}

	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	userMap := make(map[string]*entity.User)
	if len(userIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, userIDs)
		if err != nil {
			return PaginatedUsers{}, fmt.Errorf("failed to load users: %w", err)
		}
		for _, u := range users {
			userMap[u.ID] = u
		}
	}

	items := make([]workspaceuc.MemberInfo, 0, len(members))
	for _, m := range members {
		info := workspaceuc.MemberInfo{
			UserID:   m.UserID,
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt,
		}
		if user, exists := userMap[m.UserID]; exists && user != nil {
			info.Email = user.Email
			info.DisplayName = user.DisplayName
			info.AvatarURL = user.AvatarURL
		}
		items = append(items, info)
	}

	return PaginatedUsers{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: limit,
		HasMore: offset+len(items) < total,
	}, nil
}
