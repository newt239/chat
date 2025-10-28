package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// MessageUserMention holds the schema definition for the MessageUserMention entity.
type MessageUserMention struct {
	ent.Schema
}

// Fields of the MessageUserMention.
func (MessageUserMention) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageUserMention.
func (MessageUserMention) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessageUserMention.
func (MessageUserMention) Indexes() []ent.Index {
	return []ent.Index{}
}
