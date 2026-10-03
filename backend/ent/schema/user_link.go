package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserLink はプロフィールに載せるリンクです
type UserLink struct {
	ent.Schema
}

func (UserLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_link"}}
}

func (UserLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.Int("position"),
		field.String("url").
			NotEmpty(),
	}
}

func (UserLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (UserLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "position").
			Unique(),
	}
}
