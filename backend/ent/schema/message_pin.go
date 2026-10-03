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

type MessagePin struct {
	ent.Schema
}

func (MessagePin) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_pin"}}
}

func (MessagePin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}),
		field.UUID("message_id", uuid.UUID{}),
		field.UUID("pinned_by_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (MessagePin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required(),
		edge.To("pinned_by", User.Type).
			Field("pinned_by_id").
			Unique().
			Required(),
	}
}

func (MessagePin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "message_id").
			Unique(),
		index.Fields("message_id"),
	}
}
