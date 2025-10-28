package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelReadState holds the schema definition for the ChannelReadState entity.
type ChannelReadState struct {
	ent.Schema
}

// Fields of the ChannelReadState.
func (ChannelReadState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("last_read_at").
			Default(time.Now),
	}
}

// Edges of the ChannelReadState.
func (ChannelReadState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the ChannelReadState.
func (ChannelReadState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("last_read_at"),
	}
}
