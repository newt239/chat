package search

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

// SearchUseCase は検索機能のユースケースインターフェースです
type SearchUseCase interface {
	SearchWorkspace(ctx context.Context, input WorkspaceSearchInput) (*WorkspaceSearchOutput, error)
}

// NewSearchUseCase は検索ユースケースを構築します
func NewSearchUseCase(
	workspaceRepo domainrepository.WorkspaceRepository,
	channelRepo domainrepository.ChannelRepository,
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	messageOutputBuilder *messageuc.MessageOutputBuilder,
) SearchUseCase {
	return NewWorkspaceSearcher(workspaceRepo, channelRepo, messageRepo, userRepo, userGroupRepo, messageOutputBuilder)
}
