package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Workspace holds the schema definition for the Workspace entity.
type Workspace struct {
	ent.Schema
}

// Fields of the Workspace.
func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("icon_url").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Workspace.
func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("created_by", User.Type).
			Unique().
			Required(),
		edge.From("members", WorkspaceMember.Type).
			Ref("workspace"),
		edge.From("channels", Channel.Type).
			Ref("workspace"),
		edge.From("user_groups", UserGroup.Type).
			Ref("workspace"),
	}
}

// Indexes of the Workspace.
func (Workspace) Indexes() []ent.Index {
	return []ent.Index{}
}
