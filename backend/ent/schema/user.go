package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user"}}
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
		// アプリの投稿名義。ログインできず、ワークスペースのメンバーにもならない
		field.Bool("is_app").
			Default(false),
		// 公式アプリの投稿名義。この名義の投稿は誰も削除・編集できない
		field.Bool("is_official").
			Default(false),
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
		edge.To("preference", UserPreference.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("links", UserLink.Type).
			Ref("user"),
	}
}
