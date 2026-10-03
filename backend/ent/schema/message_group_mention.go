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

type MessageGroupMention struct {
	ent.Schema
}

func (MessageGroupMention) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_group_mention"}}
}

func (MessageGroupMention) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}),
		field.UUID("group_id", uuid.UUID{}),
	}
}

func (MessageGroupMention) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required(),
		edge.To("group", UserGroup.Type).
			Field("group_id").
			Unique().
			Required(),
	}
}

func (MessageGroupMention) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id"),
		index.Fields("group_id"),
	}
}
