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

// UserNote は相手ユーザーに対する自分だけに見えるニックネームとメモです
type UserNote struct {
	ent.Schema
}

func (UserNote) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_note"}}
}

func (UserNote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("nickname").
			Optional().
			Nillable(),
		field.Text("memo").
			Optional().
			Nillable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (UserNote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner", User.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("target", User.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (UserNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("owner", "target").
			Unique(),
	}
}
