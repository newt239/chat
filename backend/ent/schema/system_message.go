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

// SystemMessage holds the schema definition for the SystemMessage entity.
type SystemMessage struct {
	ent.Schema
}

func (SystemMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "system_message"}}
}

// Fields of the SystemMessage.
func (SystemMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable(),
		// kind: 種別（例: member_joined, member_added, channel_privacy_changed, channel_name_changed, channel_description_changed, message_pinned）
		field.String("kind").
			NotEmpty(),
		// payload: 種別ごとの詳細情報(JSON)
		field.JSON("payload", map[string]any{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the SystemMessage.
func (SystemMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
		// actor: 操作を行ったユーザー（不在の場合あり）
		edge.To("actor", User.Type).
			Field("actor_id").
			Unique(),
	}
}

// Indexes of the SystemMessage.
func (SystemMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("kind"),
	}
}
