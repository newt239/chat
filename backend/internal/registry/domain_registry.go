package registry

import (
	"github.com/newt239/chat/ent"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/repository"
)

// DomainRegistry はドメイン層の依存関係を管理します
type DomainRegistry struct {
	client *ent.Client
}

// NewDomainRegistry は新しいDomainRegistryを作成します
func NewDomainRegistry(client *ent.Client) *DomainRegistry {
	return &DomainRegistry{
		client: client,
	}
}

// Repositories
func (r *DomainRegistry) NewUserRepository() domainrepository.UserRepository {
	return repository.NewUserRepository(r.client)
}

func (r *DomainRegistry) NewSessionRepository() domainrepository.SessionRepository {
	return repository.NewSessionRepository(r.client)
}

func (r *DomainRegistry) NewWorkspaceRepository() domainrepository.WorkspaceRepository {
	return repository.NewWorkspaceRepository(r.client)
}

func (r *DomainRegistry) NewChannelRepository() domainrepository.ChannelRepository {
	return repository.NewChannelRepository(r.client)
}

func (r *DomainRegistry) NewChannelMemberRepository() domainrepository.ChannelMemberRepository {
	return repository.NewChannelMemberRepository(r.client)
}

func (r *DomainRegistry) NewMessageRepository() domainrepository.MessageRepository {
	return repository.NewMessageRepository(r.client)
}

func (r *DomainRegistry) NewReadStateRepository() domainrepository.ReadStateRepository {
	return repository.NewReadStateRepository(r.client)
}

func (r *DomainRegistry) NewUserGroupRepository() domainrepository.UserGroupRepository {
	return repository.NewUserGroupRepository(r.client)
}

func (r *DomainRegistry) NewMessageUserMentionRepository() domainrepository.MessageUserMentionRepository {
	return repository.NewMessageUserMentionRepository(r.client)
}

func (r *DomainRegistry) NewMessageGroupMentionRepository() domainrepository.MessageGroupMentionRepository {
	return repository.NewMessageGroupMentionRepository(r.client)
}

func (r *DomainRegistry) NewMessageLinkRepository() domainrepository.MessageLinkRepository {
	return repository.NewLinkRepository(r.client)
}

func (r *DomainRegistry) NewBookmarkRepository() domainrepository.BookmarkRepository {
	return repository.NewBookmarkRepository(r.client)
}

func (r *DomainRegistry) NewThreadRepository() domainrepository.ThreadRepository {
	return repository.NewThreadRepository(r.client)
}

func (r *DomainRegistry) NewAttachmentRepository() domainrepository.AttachmentRepository {
	return repository.NewAttachmentRepository(r.client)
}
