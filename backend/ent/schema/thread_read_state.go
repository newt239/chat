package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ThreadReadState holds the schema definition for the ThreadReadState entity.
type ThreadReadState struct {
	ent.Schema
}

// Fields of the ThreadReadState.
func (ThreadReadState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("last_read_at").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ThreadReadState.
func (ThreadReadState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("thread", Message.Type).
			Unique().
			Required(),
	}
}

// Indexes of the ThreadReadState.
func (ThreadReadState) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "thread").
			Unique(),
	}
}
