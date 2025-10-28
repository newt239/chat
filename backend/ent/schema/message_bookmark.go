package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MessageBookmark holds the schema definition for the MessageBookmark entity.
type MessageBookmark struct {
	ent.Schema
}

// Fields of the MessageBookmark.
func (MessageBookmark) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageBookmark.
func (MessageBookmark) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("message", Message.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessageBookmark.
func (MessageBookmark) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
