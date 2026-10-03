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

// MessageBookmark holds the schema definition for the MessageBookmark entity.
type MessageBookmark struct {
	ent.Schema
}

func (MessageBookmark) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_bookmark"}}
}

// Fields of the MessageBookmark.
func (MessageBookmark) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("message_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageBookmark.
func (MessageBookmark) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required(),
	}
}

// Indexes of the MessageBookmark.
func (MessageBookmark) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("user_id", "message_id").
			Unique(),
	}
}
