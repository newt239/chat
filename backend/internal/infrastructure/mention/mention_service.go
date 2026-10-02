package mention

import (
	"context"
	"fmt"
	"slices"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type mentionService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	userGroupRepo repository.UserGroupRepository
	channelRepo   repository.ChannelRepository
}

func NewMentionService(
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
	userGroupRepo repository.UserGroupRepository,
	channelRepo repository.ChannelRepository,
) service.MentionService {
	return &mentionService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
		channelRepo:   channelRepo,
	}
}

func (s *mentionService) Resolve(ctx context.Context, body, workspaceID string, knownGroups []string) (*service.ResolvedMentions, error) {
	tokens := entity.ParseMentionTokens(body)
	resolved := &service.ResolvedMentions{GroupMembers: map[string][]string{}, Channel: tokens.Channel, Here: tokens.Here}

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
