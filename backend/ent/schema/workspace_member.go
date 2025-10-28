package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// WorkspaceMember holds the schema definition for the WorkspaceMember entity.
type WorkspaceMember struct {
	ent.Schema
}

// Fields of the WorkspaceMember.
func (WorkspaceMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("role").
			NotEmpty(),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the WorkspaceMember.
func (WorkspaceMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the WorkspaceMember.
func (WorkspaceMember) Indexes() []ent.Index {
	return []ent.Index{}
}
