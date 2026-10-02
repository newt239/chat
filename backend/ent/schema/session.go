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

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "session"}}
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("refresh_token_hash").
			NotEmpty(),
		field.Time("expires_at"),
		field.Time("revoked_at").
			Optional(),
		field.String("ip_address").
			Default(""),
		field.String("user_agent").
			Default(""),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the Session.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
		index.Fields("refresh_token_hash").Unique(),
	}
}
