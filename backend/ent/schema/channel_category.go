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

// ChannelCategory はユーザーが自分のサイドバーに作るチャンネルのカテゴリです
type ChannelCategory struct {
	ent.Schema
}

func (ChannelCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_category"}}
}

func (ChannelCategory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Int("position").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (ChannelCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("workspace", Workspace.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("items", ChannelCategoryItem.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ChannelCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("position").
			Edges("user", "workspace"),
	}
}
