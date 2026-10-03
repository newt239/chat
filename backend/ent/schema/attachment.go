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

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

func (Attachment) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "attachment"}}
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("message_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("uploader_id", uuid.UUID{}),
		field.UUID("channel_id", uuid.UUID{}),
		field.String("file_name").
			NotEmpty(),
		field.String("mime_type").
			NotEmpty(),
		field.Int64("size_bytes").
			NonNegative(),
		field.Int32("width").
			Optional().
			Nillable(),
		field.Int32("height").
			Optional().
			Nillable(),
		field.Float("duration_seconds").
			Optional().
			Nillable(),
		field.String("storage_key").
			NotEmpty(),
		field.String("thumbnail_storage_key").
			Optional().
			Nillable(),
		field.Int32("thumbnail_width").
			Optional().
			Nillable(),
		field.Int32("thumbnail_height").
			Optional().
			Nillable(),
		field.String("status").
			Default("pending"),
		field.Time("uploaded_at").
			Optional(),
		field.Time("expires_at").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique(),
		edge.To("uploader", User.Type).
			Field("uploader_id").
			Unique().
			Required(),
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required(),
	}
}

// Indexes of the Attachment.
func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("message_id"),
		index.Fields("channel_id"),
	}
}
