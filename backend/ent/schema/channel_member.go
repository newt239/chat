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

type ChannelMember struct {
	ent.Schema
}

func (ChannelMember) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_member"}}
}

func (ChannelMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.String("role").
			Default("member"),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

func (ChannelMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (ChannelMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "user_id").
			Unique(),
		index.Fields("user_id"),
	}
}
