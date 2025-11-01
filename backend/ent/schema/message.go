package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Text("body").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("edited_at").
			Optional(),
		field.Time("deleted_at").
			Optional(),
		field.UUID("deleted_by", uuid.UUID{}).
			Optional(),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required(),
		edge.To("user", User.Type).
			Unique().
			Required(),
		edge.To("parent", Message.Type).
			Unique(),
		edge.From("replies", Message.Type).
			Ref("parent"),
		edge.From("reactions", MessageReaction.Type).
			Ref("message"),
		edge.From("bookmarks", MessageBookmark.Type).
			Ref("message"),
		edge.From("user_mentions", MessageUserMention.Type).
			Ref("message"),
		edge.From("group_mentions", MessageGroupMention.Type).
			Ref("message"),
		edge.From("links", MessageLink.Type).
			Ref("message"),
		edge.From("attachments", Attachment.Type).
			Ref("message"),
		edge.From("user_thread_follows", UserThreadFollow.Type).
			Ref("thread"),
		edge.From("thread_read_states", ThreadReadState.Type).
			Ref("thread"),
	}
}

// Indexes of the Message.
func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
