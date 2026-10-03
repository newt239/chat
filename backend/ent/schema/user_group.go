package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserGroup holds the schema definition for the UserGroup entity.
type UserGroup struct {
	ent.Schema
}

func (UserGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_group"}}
}

// Fields of the UserGroup.
func (UserGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id"),
		field.UUID("created_by_id", uuid.UUID{}),
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
			Field("workspace_id").
			Unique().
			Required(),
		edge.To("created_by", User.Type).
			Field("created_by_id").
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
		index.Fields("workspace_id", "name").
			Unique(),
	}
}
