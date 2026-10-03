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

type ChannelReadState struct {
	ent.Schema
}

func (ChannelReadState) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_read_state"}}
}

func (ChannelReadState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("last_read_at").
			Default(time.Now),
	}
}

func (ChannelReadState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

func (ChannelReadState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "user_id").
			Unique(),
	}
}
