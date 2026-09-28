package presenter

import (
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

func SearchResult(out *searchuc.WorkspaceSearchOutput) *chatv1.SearchWorkspaceResponse {
	return &chatv1.SearchWorkspaceResponse{
		Messages: &chatv1.MessageSearchResult{
			Items:   ConvertAll(out.Messages.Items, Message),
			Total:   int32(out.Messages.Total),
			Page:    int32(out.Messages.Page),
			PerPage: int32(out.Messages.PerPage),
			HasMore: out.Messages.HasMore,
		},
		Channels: &chatv1.ChannelSearchResult{
			Items:   ConvertAll(out.Channels.Items, Channel),
			Total:   int32(out.Channels.Total),
			Page:    int32(out.Channels.Page),
			PerPage: int32(out.Channels.PerPage),
			HasMore: out.Channels.HasMore,
		},
		Users: &chatv1.UserSearchResult{
			Items:   ConvertAll(out.Users.Items, WorkspaceMember),
			Total:   int32(out.Users.Total),
			Page:    int32(out.Users.Page),
			PerPage: int32(out.Users.PerPage),
			HasMore: out.Users.HasMore,
		},
		Groups: &chatv1.UserGroupSearchResult{
			Items:   ConvertAll(out.Groups.Items, UserGroup),
			Total:   int32(out.Groups.Total),
			Page:    int32(out.Groups.Page),
			PerPage: int32(out.Groups.PerPage),
			HasMore: out.Groups.HasMore,
		},
	}
}
