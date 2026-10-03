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

// PushToken はプッシュ通知を送る端末の FCM の送信先 ID（Firebase Installation ID）です
type PushToken struct {
	ent.Schema
}

func (PushToken) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "push_token"}}
}

func (PushToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		// 同じ端末で別のユーザーがログインしたら付け替える
		field.Text("token").
			Unique().
			NotEmpty(),
		field.Enum("platform").
			Values("web", "ios", "android"),
		field.String("user_agent").
			Default(""),
		field.Time("last_seen_at").
			Default(time.Now),
	}
}

func (PushToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (PushToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
