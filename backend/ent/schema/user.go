package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

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
		// 初回の Google ログインで紐付ける
		field.String("google_sub").
			Optional().
			Nillable().
			Unique(),
		field.String("display_name").
			NotEmpty(),
		field.String("bio").
			Optional(),
		field.String("avatar_url").
			Optional(),
		// Webhook の投稿名義。ログインできず、ワークスペースのメンバーにもならない
		field.Bool("is_bot").
			Default(false),
		// 表示設定。既定値は Jade プリセット・システムに合わせる・日本語
		field.Int("theme_hue").
			Default(168),
		field.Float("theme_chroma").
			Default(0.12),
		field.Enum("theme_sidebar").
			Values("tinted", "light").
			Default("tinted"),
		field.Enum("color_mode").
			Values("light", "dark", "system").
			Default("system"),
		field.String("locale").
			Default("ja"),
		field.Enum("notification_level").
			Values("all", "mentions", "none").
			Default("mentions"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sessions", Session.Type).
			Ref("user"),
		edge.From("created_workspaces", Workspace.Type).
			Ref("created_by"),
		edge.From("workspace_members", WorkspaceMember.Type).
			Ref("user"),
		edge.From("created_channels", Channel.Type).
			Ref("created_by"),
		edge.From("channel_members", ChannelMember.Type).
			Ref("user"),
		edge.From("messages", Message.Type).
			Ref("user"),
		edge.From("message_reactions", MessageReaction.Type).
			Ref("user"),
		edge.From("message_bookmarks", MessageBookmark.Type).
			Ref("user"),
		edge.From("user_mentions", MessageUserMention.Type).
			Ref("user"),
		edge.From("user_group_members", UserGroupMember.Type).
			Ref("user"),
		edge.From("created_user_groups", UserGroup.Type).
			Ref("created_by"),
		edge.From("attachments", Attachment.Type).
			Ref("uploader"),
		edge.From("channel_read_states", ChannelReadState.Type).
			Ref("user"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}
