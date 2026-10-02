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

// UserThreadFollow holds the schema definition for the UserThreadFollow entity.
type UserThreadFollow struct {
	ent.Schema
}

func (UserThreadFollow) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_thread_follow"}}
}

// Fields of the UserThreadFollow.
func (UserThreadFollow) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("thread_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the UserThreadFollow.
func (UserThreadFollow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("thread", Message.Type).
			Field("thread_id").
			Unique().
			Required(),
	}
}

// Indexes of the UserThreadFollow.
func (UserThreadFollow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("user_id", "thread_id").
			Unique(),
		index.Fields("thread_id"),
	}
}
