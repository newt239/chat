package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserGroupMember holds the schema definition for the UserGroupMember entity.
type UserGroupMember struct {
	ent.Schema
}

// Fields of the UserGroupMember.
func (UserGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the UserGroupMember.
func (UserGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group", UserGroup.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the UserGroupMember.
func (UserGroupMember) Indexes() []ent.Index {
	return []ent.Index{}
}
