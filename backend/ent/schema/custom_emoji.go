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

// CustomEmoji はワークスペースで登録した絵文字です。本文やリアクションでは :name: と書きます
type CustomEmoji struct {
	ent.Schema
}

func (CustomEmoji) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "custom_emoji"}}
}

func (CustomEmoji) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id").
			NotEmpty().
			Immutable(),
		field.String("name").
			NotEmpty().
			Immutable(),
		field.String("storage_key").
			NotEmpty().
			Immutable(),
		// 登録者が退会しても絵文字は残すため、ユーザーへの外部キーは張らない
		field.UUID("created_by", uuid.UUID{}).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (CustomEmoji) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (CustomEmoji) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "name").Unique(),
	}
}
