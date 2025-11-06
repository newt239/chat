package schema

import (
    "regexp"
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

// Workspace holds the schema definition for the Workspace entity.
type Workspace struct {
	ent.Schema
}

// Fields of the Workspace.
func (Workspace) Fields() []ent.Field {
	return []ent.Field{
        field.String("id").
            MaxLen(12).
            MinLen(3).
            NotEmpty().
            Unique().
            Immutable().
            Match(regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)),
        field.String("name").
            NotEmpty(),
        field.String("description").
            Optional(),
        field.String("icon_url").
            Optional(),
        field.Bool("is_public").
            Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Workspace.
func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("created_by", User.Type).
			Unique().
			Required(),
		edge.From("members", WorkspaceMember.Type).
			Ref("workspace"),
		edge.From("channels", Channel.Type).
			Ref("workspace"),
		edge.From("user_groups", UserGroup.Type).
			Ref("workspace"),
	}
}

// Indexes of the Workspace.
func (Workspace) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("id").Unique(),
        index.Fields("is_public"),
    }
}
