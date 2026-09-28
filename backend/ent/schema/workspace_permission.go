package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// WorkspacePermission はロールごとの操作権限のうち、既定値から変更されたものを保持します
type WorkspacePermission struct {
	ent.Schema
}

func (WorkspacePermission) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "workspace_permission"}}
}

func (WorkspacePermission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id").
			NotEmpty().
			Immutable(),
		field.String("role").
			NotEmpty().
			Immutable(),
		field.String("permission").
			NotEmpty().
			Immutable(),
		field.Bool("allowed"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkspacePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "role", "permission").Unique(),
	}
}
