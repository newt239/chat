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

type WorkspaceMember struct {
	ent.Schema
}

func (WorkspaceMember) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "workspace_member"}}
}

func (WorkspaceMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id"),
		field.UUID("user_id", uuid.UUID{}),
		field.String("role").
			NotEmpty(),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
		// 停止中のメンバーはワークスペースの API を利用できない
		field.Time("suspended_at").
			Optional().
			Nillable(),
	}
}

func (WorkspaceMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (WorkspaceMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "user_id").Unique(),
	}
}
