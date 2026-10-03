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

type MessageReaction struct {
	ent.Schema
}

func (MessageReaction) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_reaction"}}
}

func (MessageReaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.String("emoji").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (MessageReaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (MessageReaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id", "user_id", "emoji").
			Unique(),
		index.Fields("message_id"),
	}
}
