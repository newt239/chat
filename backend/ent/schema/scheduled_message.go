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

// ScheduledMessage は指定した日時にサーバーが代わりに投稿するメッセージです
type ScheduledMessage struct {
	ent.Schema
}

func (ScheduledMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "scheduled_message"}}
}

func (ScheduledMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Immutable(),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.Text("body"),
		// 予約時点ではメッセージに紐付けず、送信時に通常の投稿と同じく検証して添付する
		field.Strings("attachment_ids").
			Optional(),
		field.Float("location_latitude").
			Optional().
			Nillable(),
		field.Float("location_longitude").
			Optional().
			Nillable(),
		field.Float("location_accuracy").
			Optional().
			Nillable(),
		field.String("location_label").
			Optional().
			Nillable(),
		field.Time("scheduled_at"),
		// sending はワーカーが取り出して投稿している最中
		field.Enum("status").
			Values("scheduled", "sending", "sent", "failed").
			Default("scheduled"),
		field.UUID("sent_message_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("failure_reason").
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

func (ScheduledMessage) Edges() []ent.Edge {
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
		edge.To("sent_message", Message.Type).
			Field("sent_message_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (ScheduledMessage) Indexes() []ent.Index {
	return []ent.Index{
		// ワーカーが期限の来た予約を探す
		index.Fields("scheduled_at").
			Annotations(entsql.IndexWhere("status = 'scheduled'")),
		index.Fields("user_id", "scheduled_at"),
	}
}
