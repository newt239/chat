package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Channel struct {
	ent.Schema
}

func (Channel) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel"}}
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id"),
		field.UUID("created_by_id", uuid.UUID{}),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Bool("is_private").
			Default(false),
		field.String("channel_type").
			Default("public").
			Optional(),
		field.Time("archived_at").
			Optional().
			Nillable(),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
		// DM とグループ DM を参加者で一意にするキー。DM 以外は NULL
		field.String("dm_key").
			Optional().
			Nillable().
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Unique().
			Required(),
		edge.To("created_by", User.Type).
			Field("created_by_id").
			Unique().
			Required(),
		edge.From("members", ChannelMember.Type).
			Ref("channel"),
		edge.From("messages", Message.Type).
			Ref("channel"),
		edge.From("attachments", Attachment.Type).
			Ref("channel"),
		edge.From("read_states", ChannelReadState.Type).
			Ref("channel"),
		edge.To("children", Channel.Type).
			From("parent").
			Field("parent_id").
			Unique(),
	}
}

func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_private"),
		index.Fields("workspace_id"),
		index.Fields("workspace_id", "name").
			Unique().
			Annotations(entsql.IndexWhere("dm_key IS NULL")),
		index.Fields("workspace_id", "dm_key").
			Unique(),
	}
}
