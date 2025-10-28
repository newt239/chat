package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserGroup holds the schema definition for the UserGroup entity.
type UserGroup struct {
	ent.Schema
}

// Fields of the UserGroup.
func (UserGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the UserGroup.
func (UserGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Unique().
			Required(),
		edge.To("created_by", User.Type).
			Unique().
			Required(),
		edge.From("members", UserGroupMember.Type).
			Ref("group"),
		edge.From("group_mentions", MessageGroupMention.Type).
			Ref("group"),
	}
}

// Indexes of the UserGroup.
func (UserGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Unique(),
	}
}
