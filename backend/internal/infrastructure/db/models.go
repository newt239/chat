package db

import "github.com/example/chat/internal/infrastructure/database"

// Deprecated: use the models from the database package directly.
type (
	User                = database.User
	Session             = database.Session
	Workspace           = database.Workspace
	WorkspaceMember     = database.WorkspaceMember
	Channel             = database.Channel
	ChannelMember       = database.ChannelMember
	Message             = database.Message
	MessageReaction     = database.MessageReaction
	ChannelReadState    = database.ChannelReadState
	Attachment          = database.Attachment
	UserGroup           = database.UserGroup
	UserGroupMember     = database.UserGroupMember
	MessageUserMention  = database.MessageUserMention
	MessageGroupMention = database.MessageGroupMention
	MessageLink         = database.MessageLink
)
