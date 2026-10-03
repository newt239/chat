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

// UserGroupMember holds the schema definition for the UserGroupMember entity.
type UserGroupMember struct {
	ent.Schema
}

func (UserGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_group_member"}}
}

// Fields of the UserGroupMember.
func (UserGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("group_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the UserGroupMember.
func (UserGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group", UserGroup.Type).
			Field("group_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the UserGroupMember.
func (UserGroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id", "user_id").
			Unique(),
		index.Fields("user_id"),
	}
}
