package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("file_name").
			NotEmpty(),
		field.String("mime_type").
			NotEmpty(),
		field.Int64("size_bytes").
			NonNegative(),
		field.String("storage_key").
			NotEmpty(),
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
			Unique(),
		edge.To("uploader", User.Type).
			Unique().
			Required(),
		edge.To("channel", Channel.Type).
			Unique().
			Required(),
	}
}

// Indexes of the Attachment.
func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
