package search

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// SearchUseCase は検索機能のユースケースインターフェースです
type SearchUseCase interface {
	SearchWorkspace(ctx context.Context, input WorkspaceSearchInput) (*WorkspaceSearchOutput, error)
}

type searchInteractor struct {
	workspaceSearcher *WorkspaceSearcher
}

// NewSearchUseCase は検索ユースケースを構築します
func NewSearchUseCase(
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	attachmentRepo domainrepository.AttachmentRepository,
) SearchUseCase {
	return &searchInteractor{
		workspaceSearcher: NewWorkspaceSearcher(
			workspaceRepo,
			channelRepo,
			messageRepo,
			userRepo,
			userGroupRepo,
			userMentionRepo,
			groupMentionRepo,
			linkRepo,
			attachmentRepo,
		),
	}
}

func (i *searchInteractor) SearchWorkspace(ctx context.Context, input WorkspaceSearchInput) (*WorkspaceSearchOutput, error) {
	return i.workspaceSearcher.SearchWorkspace(ctx, input)
}
