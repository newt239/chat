package schema

import (
	"regexp"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Workspace struct {
	ent.Schema
}

func (Workspace) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "workspace"}}
}

func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(12).
			MinLen(3).
			NotEmpty().
			Unique().
			Immutable().
			Match(regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)),
		field.UUID("created_by_id", uuid.UUID{}),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("icon_url").
			Optional().
			Nillable(),
		field.Bool("is_public").
			Default(false),
		field.Bool("signup_enabled").
			Default(false),
		field.Bool("email_signup_enabled").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("created_by", User.Type).
			Field("created_by_id").
			Unique().
			Required(),
		edge.From("members", WorkspaceMember.Type).
			Ref("workspace"),
	}
}

func (Workspace) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_public"),
	}
}
