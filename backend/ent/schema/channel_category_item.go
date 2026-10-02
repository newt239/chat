package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelCategoryItem はチャンネルをユーザーのカテゴリに割り当てます。1 人につき 1 チャンネル 1 件です
type ChannelCategoryItem struct {
	ent.Schema
}

func (ChannelCategoryItem) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "channel_category_item"}}
}

func (ChannelCategoryItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("category_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("channel_id", uuid.UUID{}),
	}
}

func (ChannelCategoryItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", ChannelCategory.Type).
			Ref("items").
			Field("category_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ChannelCategoryItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "channel_id").
			Unique(),
	}
}
