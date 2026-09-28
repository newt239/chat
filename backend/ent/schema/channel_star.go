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

// ChannelStar はユーザーごとのチャンネル・DM へのスターです
type ChannelStar struct {
	ent.Schema
}

func (ChannelStar) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_star"}}
}

func (ChannelStar) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (ChannelStar) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("channel", Channel.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ChannelStar) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "channel").
			Unique(),
	}
}
