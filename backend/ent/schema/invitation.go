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

// Invitation はメールアドレス宛てのワークスペースへの招待です。トークンはハッシュで保存します
type Invitation struct {
	ent.Schema
}

func (Invitation) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "invitation"}}
}

func (Invitation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			NotEmpty(),
		field.String("role").
			NotEmpty(),
		field.String("token_hash").
			NotEmpty().
			Unique().
			Sensitive(),
		field.Time("expires_at"),
		field.Time("accepted_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Invitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("invited_by", User.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Invitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}
