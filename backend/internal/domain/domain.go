package domain

import (
	"github.com/example/chat/internal/domain/entity"
	domainrepository "github.com/example/chat/internal/domain/repository"
)

type (
	Attachment          = entity.Attachment
	Channel             = entity.Channel
	ChannelMember       = entity.ChannelMember
	ChannelReadState    = entity.ChannelReadState
	Message             = entity.Message
	MessageReaction     = entity.MessageReaction
	MessageGroupMention = entity.MessageGroupMention
	MessageLink         = entity.MessageLink
	MessageUserMention  = entity.MessageUserMention
	Session             = entity.Session
	User                = entity.User
	UserGroup           = entity.UserGroup
	UserGroupMember     = entity.UserGroupMember
	Workspace           = entity.Workspace
	WorkspaceMember     = entity.WorkspaceMember
	WorkspaceRole       = entity.WorkspaceRole
)

const (
	WorkspaceRoleOwner  = entity.WorkspaceRoleOwner
	WorkspaceRoleAdmin  = entity.WorkspaceRoleAdmin
	WorkspaceRoleMember = entity.WorkspaceRoleMember
	WorkspaceRoleGuest  = entity.WorkspaceRoleGuest
)

type (
	AttachmentRepository          = domainrepository.AttachmentRepository
	ChannelRepository             = domainrepository.ChannelRepository
	MessageRepository             = domainrepository.MessageRepository
	MessageGroupMentionRepository = domainrepository.MessageGroupMentionRepository
	MessageLinkRepository         = domainrepository.MessageLinkRepository
	MessageUserMentionRepository  = domainrepository.MessageUserMentionRepository
	ReadStateRepository           = domainrepository.ReadStateRepository
	SessionRepository             = domainrepository.SessionRepository
	UserRepository                = domainrepository.UserRepository
	UserGroupRepository           = domainrepository.UserGroupRepository
	WorkspaceRepository           = domainrepository.WorkspaceRepository
)
