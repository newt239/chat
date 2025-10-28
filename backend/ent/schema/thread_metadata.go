package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ThreadMetadata holds the schema definition for the ThreadMetadata entity.
type ThreadMetadata struct {
	ent.Schema
}

// Fields of the ThreadMetadata.
func (ThreadMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("reply_count").
			Default(0).
			NonNegative(),
		field.Time("last_reply_at").
			Optional(),
		field.JSON("participant_user_ids", []uuid.UUID{}).
			Default([]uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ThreadMetadata.
func (ThreadMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Unique().
			Required(),
		edge.To("last_reply_user", User.Type).
			Unique(),
	}
}

// Indexes of the ThreadMetadata.
func (ThreadMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("last_reply_at"),
	}
}
