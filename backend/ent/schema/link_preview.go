package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// LinkPreview は URL ごとに取得した OGP です。メッセージをまたいで共有する
type LinkPreview struct {
	ent.Schema
}

func (LinkPreview) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "link_preview"}}
}

func (LinkPreview) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("url").
			NotEmpty().
			Unique(),
		field.String("title").
			Optional().
			Nillable(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("image_url").
			Optional().
			Nillable(),
		field.String("site_name").
			Optional().
			Nillable(),
		field.String("card_type").
			Optional().
			Nillable(),
		field.Int32("image_width").
			Optional().
			Nillable(),
		field.Int32("image_height").
			Optional().
			Nillable(),
		field.Time("fetched_at"),
	}
}

func (LinkPreview) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("youtube", LinkPreviewYoutube.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("x_post", LinkPreviewXPost.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("message_links", MessageLink.Type).
			Ref("link_preview"),
	}
}

type LinkPreviewYoutube struct {
	ent.Schema
}

func (LinkPreviewYoutube) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "link_preview_youtube"}}
}

func (LinkPreviewYoutube) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("link_preview_id", uuid.UUID{}).
			Unique().
			Immutable(),
		field.String("video_id").
			NotEmpty(),
		field.String("channel_name").
			Optional().
			Nillable(),
		field.Int32("duration_seconds").
			Optional().
			Nillable(),
	}
}

func (LinkPreviewYoutube) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("link_preview", LinkPreview.Type).
			Ref("youtube").
			Field("link_preview_id").
			Unique().
			Required().
			Immutable(),
	}
}

type LinkPreviewXPost struct {
	ent.Schema
}

func (LinkPreviewXPost) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "link_preview_x_post"}}
}

func (LinkPreviewXPost) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("link_preview_id", uuid.UUID{}).
			Unique().
			Immutable(),
		field.String("author_name"),
		field.String("author_handle"),
	}
}

func (LinkPreviewXPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("link_preview", LinkPreview.Type).
			Ref("x_post").
			Field("link_preview_id").
			Unique().
			Required().
			Immutable(),
	}
}
