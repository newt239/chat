package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password_hash").
			NotEmpty(),
		field.String("display_name").
			NotEmpty(),
		field.String("avatar_url").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// Sessions
		edge.From("sessions", Session.Type).
			Ref("user"),
		// Workspaces created by user
		edge.From("created_workspaces", Workspace.Type).
			Ref("created_by"),
		// Workspace memberships
		edge.From("workspace_members", WorkspaceMember.Type).
			Ref("user"),
		// Channels created by user
		edge.From("created_channels", Channel.Type).
			Ref("created_by"),
		// Channel memberships
		edge.From("channel_members", ChannelMember.Type).
			Ref("user"),
		// Messages sent by user
		edge.From("messages", Message.Type).
			Ref("user"),
		// Message reactions by user
		edge.From("message_reactions", MessageReaction.Type).
			Ref("user"),
		// Message bookmarks by user
		edge.From("message_bookmarks", MessageBookmark.Type).
			Ref("user"),
		// User mentions in messages
		edge.From("user_mentions", MessageUserMention.Type).
			Ref("user"),
		// User group memberships
		edge.From("user_group_members", UserGroupMember.Type).
			Ref("user"),
		// User groups created by user
		edge.From("created_user_groups", UserGroup.Type).
			Ref("created_by"),
		// Attachments uploaded by user
		edge.From("attachments", Attachment.Type).
			Ref("uploader"),
		// Channel read states
		edge.From("channel_read_states", ChannelReadState.Type).
			Ref("user"),
		// Thread metadata last reply user
		edge.From("thread_metadata_last_reply", ThreadMetadata.Type).
			Ref("last_reply_user"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}
