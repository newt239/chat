package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Webhook はチャンネルへの着信 Webhook です。投稿は専用のボットユーザー名義で行います
type Webhook struct {
	ent.Schema
}

func (Webhook) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "webhook"}}
}

func (Webhook) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		field.String("token_hash").
			NotEmpty().
			Sensitive(),
		field.Time("last_used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Webhook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("created_by", User.Type).
			Unique().
			Required(),
		// 削除後も過去の投稿の名義として残すため、ボットユーザーは消さない
		edge.To("bot_user", User.Type).
			Unique().
			Required(),
	}
}
