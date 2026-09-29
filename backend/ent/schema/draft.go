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

// Draft は入力途中のメッセージです。チャンネルまたはスレッドごとに 1 件だけ持ちます
type Draft struct {
	ent.Schema
}

func (Draft) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "draft"}}
}

func (Draft) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Immutable(),
		// スレッドへの返信の下書きのときだけ設定する
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.Text("body"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Draft) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("parent", Message.Type).
			Field("parent_id").
			Unique().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// NULL 同士は一意制約で区別されるため、チャンネルとスレッドの下書きを部分インデックスで分けて一意にする
func (Draft) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "channel_id").
			Unique().
			Annotations(entsql.IndexWhere("parent_id IS NULL")),
		index.Fields("user_id", "channel_id", "parent_id").
			Unique().
			Annotations(entsql.IndexWhere("parent_id IS NOT NULL")),
	}
}
