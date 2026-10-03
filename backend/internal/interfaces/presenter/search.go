package presenter

import (
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	searchuc "github.com/newt239/chat/internal/usecase/search"
)

func SearchResult(out *searchuc.WorkspaceSearchOutput) *chatv1.SearchWorkspaceResponse {
	return &chatv1.SearchWorkspaceResponse{
		Messages: &chatv1.MessageSearchResult{
			Items: ConvertAll(out.Messages.Items, messageSearchHit),
			Total: int32(out.Messages.Total),
		},
		Channels: &chatv1.ChannelSearchResult{
			Items: ConvertAll(out.Channels.Items, Channel),
			Total: int32(out.Channels.Total),
		},
		Users: &chatv1.UserSearchResult{
			Items: ConvertAll(out.Users.Items, WorkspaceMember),
			Total: int32(out.Users.Total),
		},
		Groups: &chatv1.UserGroupSearchResult{
			Items: ConvertAll(out.Groups.Items, UserGroup),
			Total: int32(out.Groups.Total),
		},
	}
}

func messageSearchHit(hit searchuc.MessageHit) *chatv1.MessageSearchHit {
	return &chatv1.MessageSearchHit{
		Message: Message(hit.Message),
		Highlights: ConvertAll(hit.Highlights, func(r searchuc.TextRange) *chatv1.TextRange {
			return &chatv1.TextRange{Start: int32(r.Start), End: int32(r.End)}
		}),
	}
}
