package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelMember holds the schema definition for the ChannelMember entity.
type ChannelMember struct {
	ent.Schema
}

// Fields of the ChannelMember.
func (ChannelMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("role").
			Default("member"),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ChannelMember.
func (ChannelMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
	}
}

// Indexes of the ChannelMember.
func (ChannelMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role"),
	}
}
