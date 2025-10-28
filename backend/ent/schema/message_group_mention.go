package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// MessageGroupMention holds the schema definition for the MessageGroupMention entity.
type MessageGroupMention struct {
	ent.Schema
}

// Fields of the MessageGroupMention.
func (MessageGroupMention) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageGroupMention.
func (MessageGroupMention) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Unique().
			Required(),
		edge.To("group", UserGroup.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessageGroupMention.
func (MessageGroupMention) Indexes() []ent.Index {
	return []ent.Index{}
}
