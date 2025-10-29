package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MessagePin holds the schema definition for the MessagePin entity.
type MessagePin struct {
	ent.Schema
}

// Fields of the MessagePin.
func (MessagePin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessagePin.
func (MessagePin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required(),
		edge.To("message", Message.Type).
			Unique().
			Required(),
		edge.To("pinned_by", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessagePin.
func (MessagePin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		// channel + message のユニーク制約
		index.Edges("channel", "message").
			Unique(),
	}
}
