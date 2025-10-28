package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Channel holds the schema definition for the Channel entity.
type Channel struct {
	ent.Schema
}

// Fields of the Channel.
func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Bool("is_private").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Channel.
func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Unique().
			Required(),
		edge.To("created_by", User.Type).
			Unique().
			Required(),
		edge.From("members", ChannelMember.Type).
			Ref("channel"),
		edge.From("messages", Message.Type).
			Ref("channel"),
		edge.From("attachments", Attachment.Type).
			Ref("channel"),
		edge.From("read_states", ChannelReadState.Type).
			Ref("channel"),
	}
}

// Indexes of the Channel.
func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_private"),
		index.Fields("name").
			Unique(),
	}
}
