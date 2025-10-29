package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserThreadFollow holds the schema definition for the UserThreadFollow entity.
type UserThreadFollow struct {
	ent.Schema
}

// Fields of the UserThreadFollow.
func (UserThreadFollow) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the UserThreadFollow.
func (UserThreadFollow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("thread", Message.Type).
			Unique().
			Required(),
	}
}

// Indexes of the UserThreadFollow.
func (UserThreadFollow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Edges("user", "thread").
			Unique(),
	}
}
