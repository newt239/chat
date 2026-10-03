package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// ResolvedMentions は本文の ID 記法のうち、ワークスペースに実在するものを投稿時点で解決した結果です
type ResolvedMentions struct {
	UserIDs  []string
	GroupIDs []string
	// グループ ID ごとの、解決した時点のメンバー
	GroupMembers map[string][]string
}

type MentionService interface {
	// Resolve は本文の ID 記法を検証して解決します。knownGroups に含まれるグループは展開せず、呼び出し側が以前の展開結果を使います
	Resolve(ctx context.Context, body, workspaceID string, knownGroups []string) (*ResolvedMentions, error)
	// RenderPlain は本文の ID 記法を現在の名前に置き換えた文字列を返します
	RenderPlain(ctx context.Context, body string) (string, error)
}

type mentionService struct {
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
	userGroupRepo domainrepository.UserGroupRepository
	channelRepo   domainrepository.ChannelRepository
}

func NewMentionService(
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	channelRepo domainrepository.ChannelRepository,
) MentionService {
	return &mentionService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
		channelRepo:   channelRepo,
	}
}

func (s *mentionService) Resolve(ctx context.Context, body, workspaceID string, knownGroups []string) (*ResolvedMentions, error) {
	tokens := entity.ParseMentionTokens(body)
	resolved := &ResolvedMentions{GroupMembers: map[string][]string{}}

	if len(tokens.UserIDs) > 0 {
		members, err := s.workspaceRepo.FindActiveMemberIDs(ctx, workspaceID, tokens.UserIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load workspace members: %w", err)
		}
		for _, userID := range tokens.UserIDs {
			if members[userID] {
				resolved.UserIDs = append(resolved.UserIDs, userID)
			}
		}
	}

	if len(tokens.GroupIDs) > 0 {
		groups, err := s.userGroupRepo.FindByIDs(ctx, tokens.GroupIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load groups: %w", err)
		}
		var expand []string
		for _, group := range groups {
			if group.WorkspaceID != workspaceID {
				continue
			}
			resolved.GroupIDs = append(resolved.GroupIDs, group.ID)
			if !slices.Contains(knownGroups, group.ID) {
				expand = append(expand, group.ID)
			}
		}
		if len(expand) > 0 {
			members, err := s.userGroupRepo.FindMembersByGroupIDs(ctx, expand)
			if err != nil {
				return nil, fmt.Errorf("failed to load group members: %w", err)
			}
			for _, member := range members {
				resolved.GroupMembers[member.GroupID] = append(resolved.GroupMembers[member.GroupID], member.UserID)
			}
		}
	}
	return resolved, nil
}

func (s *mentionService) RenderPlain(ctx context.Context, body string) (string, error) {
	tokens := entity.ParseMentionTokens(body)
	names := map[string]string{}
	if len(tokens.UserIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, tokens.UserIDs)
		if err != nil {
			return "", err
		}
		for _, u := range users {
			names[u.ID] = u.DisplayName
		}
	}
	if len(tokens.GroupIDs) > 0 {
		groups, err := s.userGroupRepo.FindByIDs(ctx, tokens.GroupIDs)
		if err != nil {
			return "", err
		}
		for _, g := range groups {
			names[g.ID] = g.Name
		}
	}
	if len(tokens.ChannelIDs) > 0 {
		channels, err := s.channelRepo.FindByIDs(ctx, tokens.ChannelIDs)
		if err != nil {
			return "", err
		}
		for _, c := range channels {
			names[c.ID] = c.Name
		}
	}
	return entity.ReplaceMentionTokens(body, func(kind entity.MentionKind, id string) string {
		prefix := "@"
		if kind == entity.MentionKindChannel {
			prefix = "#"
		}
		if kind == entity.MentionKindBroadcast {
			return prefix + id
		}
		if name, ok := names[id]; ok {
			return prefix + name
		}
		return prefix + "unknown"
	}), nil
}
