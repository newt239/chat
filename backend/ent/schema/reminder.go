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

// Reminder は /remind で設定したリマインダーです。指定日時に公式アプリが届ける
type Reminder struct {
	ent.Schema
}

func (Reminder) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "reminder"}}
}

func (Reminder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id").
			Immutable(),
		field.UUID("creator_id", uuid.UUID{}).
			Immutable(),
		// どちらもなければ作成者本人に届ける
		field.UUID("target_user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.UUID("target_channel_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.Text("text").
			NotEmpty(),
		field.Time("remind_at"),
		field.Enum("status").
			Values("scheduled", "sending", "sent", "failed").
			Default("scheduled"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Reminder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("creator", User.Type).
			Field("creator_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Reminder) Indexes() []ent.Index {
	return []ent.Index{
		// 期限の来たものを取り出す
		index.Fields("remind_at").
			Annotations(entsql.IndexWhere("status = 'scheduled'")),
	}
}
