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

// MessageUserMention holds the schema definition for the MessageUserMention entity.
type MessageUserMention struct {
	ent.Schema
}

func (MessageUserMention) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_user_mention"}}
}

// Fields of the MessageUserMention.
func (MessageUserMention) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		// グループへのメンションを投稿時点のメンバーに展開した行の展開元
		field.UUID("via_group_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageUserMention.
func (MessageUserMention) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the MessageUserMention.
func (MessageUserMention) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id"),
		index.Fields("user_id"),
	}
}
