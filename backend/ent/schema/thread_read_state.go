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

type ThreadReadState struct {
	ent.Schema
}

func (ThreadReadState) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "thread_read_state"}}
}

func (ThreadReadState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("thread_id", uuid.UUID{}),
		field.Time("last_read_at").
			Default(time.Now),
	}
}

func (ThreadReadState) Edges() []ent.Edge {
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

func (ThreadReadState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "thread_id").
			Unique(),
	}
}
