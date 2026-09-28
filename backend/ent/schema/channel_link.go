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

// ChannelLink はチャンネルの関連リンクです
type ChannelLink struct {
	ent.Schema
}

func (ChannelLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_link"}}
}

func (ChannelLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty(),
		field.String("url").
			NotEmpty(),
		field.Int("position").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (ChannelLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("created_by", User.Type).
			Unique().
			Required(),
	}
}

func (ChannelLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("position").
			Edges("channel"),
	}
}
