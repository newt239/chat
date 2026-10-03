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

type SystemMessage struct {
	ent.Schema
}

func (SystemMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "system_message"}}
}

func (SystemMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("kind").
			NotEmpty(),
		field.JSON("payload", map[string]any{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (SystemMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
		edge.To("actor", User.Type).
			Field("actor_id").
			Unique(),
	}
}

func (SystemMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "created_at"),
	}
}
