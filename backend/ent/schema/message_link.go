package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MessageLink holds the schema definition for the MessageLink entity.
type MessageLink struct {
	ent.Schema
}

// Fields of the MessageLink.
func (MessageLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("url").
			NotEmpty(),
		field.String("title").
			Optional(),
		field.String("description").
			Optional(),
		field.String("image_url").
			Optional(),
		field.String("site_name").
			Optional(),
		field.String("card_type").
			Optional(),
		field.Int32("image_width").
			Optional().
			Nillable(),
		field.Int32("image_height").
			Optional().
			Nillable(),
		field.String("youtube_video_id").
			Optional().
			Nillable(),
		field.String("youtube_channel_name").
			Optional().
			Nillable(),
		field.Int32("youtube_duration_seconds").
			Optional().
			Nillable(),
		field.String("x_author_name").
			Optional().
			Nillable(),
		field.String("x_author_handle").
			Optional().
			Nillable(),
		field.UUID("linked_message_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the MessageLink.
func (MessageLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Unique().
			Required(),
	}
}

// Indexes of the MessageLink.
func (MessageLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("url").
			Edges("message").
			Unique(),
		index.Edges("message"),
	}
}
