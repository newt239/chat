package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MessageReaction holds the schema definition for the MessageReaction entity.
type MessageReaction struct {
	ent.Schema
}

// Fields of the MessageReaction.
func (MessageReaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("emoji").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageReaction.
func (MessageReaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessageReaction.
func (MessageReaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("emoji").
			Edges("message", "user").
			Unique(),
		index.Edges("message"),
	}
}
