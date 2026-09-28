package search

import (
	"context"
	"fmt"
	"slices"
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
	maxTerms       = 10
)

type WorkspaceSearcher struct {
	workspaceRepo        domainrepository.WorkspaceRepository
	channelRepo          domainrepository.ChannelRepository
	messageRepo          domainrepository.MessageRepository
	searchIndex          domainrepository.MessageSearchIndex
	userRepo             domainrepository.UserRepository
	userGroupRepo        domainrepository.UserGroupRepository
	messageOutputBuilder *messageuc.MessageOutputBuilder
}

func NewWorkspaceSearcher(
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	messageRepo domainrepository.MessageRepository,
	searchIndex domainrepository.MessageSearchIndex,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	messageOutputBuilder *messageuc.MessageOutputBuilder,
) *WorkspaceSearcher {
	return &WorkspaceSearcher{
		workspaceRepo:        workspaceRepo,
		channelRepo:          channelRepo,
		messageRepo:          messageRepo,
		searchIndex:          searchIndex,
		userRepo:             userRepo,
		userGroupRepo:        userGroupRepo,
		messageOutputBuilder: messageOutputBuilder,
	}
}

func (s *WorkspaceSearcher) SearchWorkspace(ctx context.Context, input WorkspaceSearchInput) (*WorkspaceSearchOutput, error) {
	terms := splitTerms(input.Query)
	if len(terms) == 0 && input.Filter.isEmpty() {
		return nil, ErrInvalidQuery
	}
	if input.Filter.After != nil && input.Filter.Before != nil && !input.Filter.After.Before(*input.Filter.Before) {
		return nil, ErrInvalidDateRange
	}

	target := input.Target.Normalize()
	page := max(input.Page, 1)
	perPage := input.PerPage
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	perPage = min(perPage, maxPerPage)
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

	out := &WorkspaceSearchOutput{
		Messages: PaginatedMessages{Items: []MessageHit{}, Page: page, PerPage: perPage},
		Channels: PaginatedChannels{Items: []channeluc.ChannelOutput{}, Page: page, PerPage: perPage},
		Users:    PaginatedUsers{Items: []workspaceuc.MemberInfo{}, Page: page, PerPage: perPage},
		Groups:   PaginatedUserGroups{Items: []usergroupuc.UserGroupOutput{}, Page: page, PerPage: perPage},
	}

	if target.includesMessages() {
		if out.Messages, err = s.searchMessages(ctx, input, terms, page, perPage, offset); err != nil {
			return nil, err
		}
	}

	// チャンネル・ユーザー・グループはキーワードでのみ検索する
	keyword := strings.Join(terms, " ")
	if keyword == "" {
		return out, nil
	}
	if target.includesChannels() {
		if out.Channels, err = s.searchChannels(ctx, keyword, input.WorkspaceID, input.RequesterID, page, perPage, offset); err != nil {
			return nil, err
		}
	}
	if target.includesUsers() {
		if out.Users, err = s.searchUsers(ctx, keyword, input.WorkspaceID, page, perPage, offset); err != nil {
			return nil, err
		}
	}
	if target.includesGroups() {
		if out.Groups, err = s.searchUserGroups(ctx, keyword, input.WorkspaceID, page, perPage, offset); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *WorkspaceSearcher) searchMessages(
	ctx context.Context,
	input WorkspaceSearchInput,
	terms []string,
	page int,
	limit int,
	offset int,
) (PaginatedMessages, error) {
	result := PaginatedMessages{Items: []MessageHit{}, Page: page, PerPage: limit}
	f := input.Filter

	channelIDs := f.ChannelIDs
	if len(channelIDs) > 0 && f.IncludeDescendantChannels {
		channels, err := s.channelRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
		if err != nil {
			return PaginatedMessages{}, fmt.Errorf("failed to load channels: %w", err)
		}
		channelIDs = withDescendantChannelIDs(channels, channelIDs)
	}

	scope, err := s.messageRepo.FindSearchScope(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to load search scope: %w", err)
	}
	channelIDs = searchableChannelIDs(scope.ViewableChannelIDs, channelIDs)
	if len(channelIDs) == 0 {
		return result, nil
	}

	criteria := domainrepository.MessageSearchCriteria{
		WorkspaceID:    input.WorkspaceID,
		Terms:          terms,
		ChannelIDs:     channelIDs,
		AuthorIDs:      f.FromUserIDs,
		Has:            f.Has,
		PinnedOnly:     f.PinnedOnly,
		ThreadOnly:     f.ThreadOnly,
		ExcludeReplies: f.ExcludeReplies,
		After:          f.After,
		Before:         f.Before,
		Sort:           input.Sort,
		Page:           page,
		PerPage:        limit,
	}
	if f.MentionsMe {
		criteria.Mention = scope
	}
	// キーワードがなければ関連度は決まらないため新しい順にする
	if len(terms) == 0 {
		criteria.Sort = domainrepository.MessageSearchSortNewest
	}
	hits, err := s.searchIndex.Search(ctx, criteria)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to search messages: %w", err)
	}
	messages, err := s.findLiveMessages(ctx, hits.MessageIDs)
	if err != nil {
		return PaginatedMessages{}, err
	}

	outputs, err := s.messageOutputBuilder.Build(ctx, input.RequesterID, messages)
	if err != nil {
		return PaginatedMessages{}, fmt.Errorf("failed to build message outputs: %w", err)
	}
	for _, o := range outputs {
		result.Items = append(result.Items, MessageHit{Message: o, Highlights: highlightRanges(o.Body, terms)})
	}
	result.Total = hits.Total
	result.HasMore = offset+len(hits.MessageIDs) < hits.Total
	return result, nil
}

// findLiveMessages はインデックスが返した順に、削除されていないメッセージを読み込みます
func (s *WorkspaceSearcher) findLiveMessages(ctx context.Context, ids []string) ([]*entity.Message, error) {
	found, err := s.messageRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	byID := make(map[string]*entity.Message, len(found))
	for _, m := range found {
		byID[m.ID] = m
	}
	messages := make([]*entity.Message, 0, len(ids))
	for _, id := range ids {
		if m := byID[id]; m != nil && m.DeletedAt == nil {
			messages = append(messages, m)
		}
	}
	return messages, nil
}

// searchableChannelIDs は閲覧できるチャンネルのうち、指定があればそれに含まれるものを返します
func searchableChannelIDs(viewable []string, requested []string) []string {
	if len(requested) == 0 {
		return viewable
	}
	result := []string{}
	for _, id := range requested {
		if slices.Contains(viewable, id) {
			result = append(result, id)
		}
	}
	return result
}

// splitTerms は全角を含む空白で区切った語を重複なく最大 maxTerms 個返します
func splitTerms(query string) []string {
	terms := []string{}
	seen := map[string]bool{}
	for _, term := range strings.Fields(query) {
		key := strings.ToLower(term)
		if seen[key] {
			continue
		}
		seen[key] = true
		terms = append(terms, term)
		if len(terms) == maxTerms {
			break
		}
	}
	return terms
}

// withDescendantChannelIDs は指定したチャンネルと、名前が "<親の名前>/" で始まる下階層のチャンネルの ID を返します
func withDescendantChannelIDs(channels []*entity.Channel, ids []string) []string {
	selected := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	prefixes := []string{}
	for _, ch := range channels {
		if selected[ch.ID] {
			prefixes = append(prefixes, ch.Name+"/")
		}
	}
	result := append([]string{}, ids...)
	for _, ch := range channels {
		if selected[ch.ID] {
			continue
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(ch.Name, prefix) {
				result = append(result, ch.ID)
				break
			}
		}
	}
	return result
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
