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

// Poll はメッセージに付けた投票です
type Poll struct {
	ent.Schema
}

func (Poll) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "poll"}}
}

func (Poll) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}).
			Unique().
			Immutable(),
		field.Text("question").
			NotEmpty(),
		// text は文字の選択肢、date は日時の候補から選ぶ日程調整
		field.Enum("mode").
			Values("text", "date").
			Immutable(),
		field.Bool("allow_multiple").
			Immutable(),
		// 誰がどれに投票したかを作成者にも見せない
		field.Bool("anonymous").
			Immutable(),
		field.Time("closes_at").
			Optional().
			Nillable().
			Immutable(),
		field.Time("closed_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Poll) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique().
			Required().
			Immutable(),
		edge.From("options", PollOption.Type).
			Ref("poll"),
	}
}

// PollOption は投票の選択肢です
type PollOption struct {
	ent.Schema
}

func (PollOption) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "poll_option"}}
}

func (PollOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}).
			Immutable(),
		field.Int("position").
			Immutable(),
		field.String("label").
			Default("").
			Immutable(),
		// 日程調整の候補。all_day なら日付だけを使う
		field.Time("starts_at").
			Optional().
			Nillable().
			Immutable(),
		field.Bool("all_day").
			Default(false).
			Immutable(),
	}
}

func (PollOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("poll", Poll.Type).
			Field("poll_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (PollOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("poll_id", "position"),
	}
}

// PollVote は選択肢への 1 票です
type PollVote struct {
	ent.Schema
}

func (PollVote) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "poll_vote"}}
}

func (PollVote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("option_id", uuid.UUID{}).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (PollVote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("option", PollOption.Type).
			Field("option_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (PollVote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("option_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
