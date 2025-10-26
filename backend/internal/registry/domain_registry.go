package registry

import (
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/repository"
	"gorm.io/gorm"
)

// DomainRegistry はドメイン層の依存関係を管理します
type DomainRegistry struct {
	db *gorm.DB
}

// NewDomainRegistry は新しいDomainRegistryを作成します
func NewDomainRegistry(db *gorm.DB) *DomainRegistry {
	return &DomainRegistry{
		db: db,
	}
}

// Repositories
func (r *DomainRegistry) NewUserRepository() domainrepository.UserRepository {
	return repository.NewUserRepository(r.db)
}

func (r *DomainRegistry) NewSessionRepository() domainrepository.SessionRepository {
	return repository.NewSessionRepository(r.db)
}

func (r *DomainRegistry) NewWorkspaceRepository() domainrepository.WorkspaceRepository {
	return repository.NewWorkspaceRepository(r.db)
}

func (r *DomainRegistry) NewChannelRepository() domainrepository.ChannelRepository {
	return repository.NewChannelRepository(r.db)
}

func (r *DomainRegistry) NewChannelMemberRepository() domainrepository.ChannelMemberRepository {
	return repository.NewChannelMemberRepository(r.db)
}

func (r *DomainRegistry) NewMessageRepository() domainrepository.MessageRepository {
	return repository.NewMessageRepository(r.db)
}

func (r *DomainRegistry) NewReadStateRepository() domainrepository.ReadStateRepository {
	return repository.NewReadStateRepository(r.db)
}

func (r *DomainRegistry) NewUserGroupRepository() domainrepository.UserGroupRepository {
	return repository.NewUserGroupRepository(r.db)
}

func (r *DomainRegistry) NewMessageUserMentionRepository() domainrepository.MessageUserMentionRepository {
	return repository.NewMessageUserMentionRepository(r.db)
}

func (r *DomainRegistry) NewMessageGroupMentionRepository() domainrepository.MessageGroupMentionRepository {
	return repository.NewMessageGroupMentionRepository(r.db)
}

func (r *DomainRegistry) NewMessageLinkRepository() domainrepository.MessageLinkRepository {
	return repository.NewMessageLinkRepository(r.db)
}

func (r *DomainRegistry) NewBookmarkRepository() domainrepository.BookmarkRepository {
	return repository.NewBookmarkRepository(r.db)
}

func (r *DomainRegistry) NewThreadRepository() domainrepository.ThreadRepository {
	return repository.NewThreadRepository(r.db)
}

func (r *DomainRegistry) NewAttachmentRepository() domainrepository.AttachmentRepository {
	return repository.NewAttachmentRepository(r.db)
}
