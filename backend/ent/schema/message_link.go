package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
			NotEmpty().
			Unique(),
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
	return []ent.Index{}
}
