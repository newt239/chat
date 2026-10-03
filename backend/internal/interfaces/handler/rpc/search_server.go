package rpc

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

type SearchServer struct {
	UC searchuc.SearchUseCase
}

var searchTargets = map[chatv1.SearchTarget]searchuc.SearchTarget{
	chatv1.SearchTarget_SEARCH_TARGET_ALL:      searchuc.SearchTargetAll,
	chatv1.SearchTarget_SEARCH_TARGET_MESSAGES: searchuc.SearchTargetMessages,
	chatv1.SearchTarget_SEARCH_TARGET_CHANNELS: searchuc.SearchTargetChannels,
	chatv1.SearchTarget_SEARCH_TARGET_USERS:    searchuc.SearchTargetUsers,
	chatv1.SearchTarget_SEARCH_TARGET_GROUPS:   searchuc.SearchTargetGroups,
}

var searchHas = map[chatv1.SearchHas]domainrepository.MessageContentKind{
	chatv1.SearchHas_SEARCH_HAS_IMAGE:    domainrepository.MessageContentImage,
	chatv1.SearchHas_SEARCH_HAS_FILE:     domainrepository.MessageContentFile,
	chatv1.SearchHas_SEARCH_HAS_LINK:     domainrepository.MessageContentLink,
	chatv1.SearchHas_SEARCH_HAS_VIDEO:    domainrepository.MessageContentVideo,
	chatv1.SearchHas_SEARCH_HAS_LOCATION: domainrepository.MessageContentLocation,
}

func (s *SearchServer) SearchWorkspace(ctx context.Context, req *chatv1.SearchWorkspaceRequest) (*chatv1.SearchWorkspaceResponse, error) {
	sort := domainrepository.MessageSearchSortNewest
	if req.Sort == chatv1.SearchSort_SEARCH_SORT_RELEVANCE {
		sort = domainrepository.MessageSearchSortRelevance
	}
	out, err := s.UC.SearchWorkspace(ctx, searchuc.WorkspaceSearchInput{
		WorkspaceID: req.WorkspaceId,
		RequesterID: userIDFrom(ctx),
		Query:       req.Query,
		Target:      searchTargets[req.Target],
		Filter:      messageFilter(req.MessageFilter),
		Sort:        sort,
		Page:        int(req.Page),
		PerPage:     int(req.PerPage),
	})
	if err != nil {
		return nil, err
	}
	return presenter.SearchResult(out), nil
}

func messageFilter(f *chatv1.MessageSearchFilter) searchuc.MessageFilter {
	if f == nil {
		return searchuc.MessageFilter{}
	}
	has := make([]domainrepository.MessageContentKind, 0, len(f.Has))
	for _, h := range f.Has {
		if kind, ok := searchHas[h]; ok {
			has = append(has, kind)
		}
	}
	return searchuc.MessageFilter{
		FromUserIDs:               f.FromUserIds,
		ChannelIDs:                f.ChannelIds,
		IncludeDescendantChannels: f.IncludeDescendantChannels,
		Has:                       has,
		PinnedOnly:                f.PinnedOnly,
		ThreadOnly:                f.ThreadOnly,
		MentionsMe:                f.MentionsMe,
		ExcludeReplies:            f.ExcludeReplies,
		After:                     optionalTime(f.After),
		Before:                    optionalTime(f.Before),
	}
}
