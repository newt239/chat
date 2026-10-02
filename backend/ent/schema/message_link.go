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

// MessageLink はメッセージ本文に含まれる URL です。OGP は link_preview に、同じワークスペースのメッセージへのリンクは linked_message_id に持つ
type MessageLink struct {
	ent.Schema
}

func (MessageLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "message_link"}}
}

func (MessageLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}).
			Immutable(),
		field.String("url").
			NotEmpty(),
		field.UUID("link_preview_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("linked_message_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (MessageLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("link_preview", LinkPreview.Type).
			Field("link_preview_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (MessageLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id", "url").
			Unique(),
	}
}
