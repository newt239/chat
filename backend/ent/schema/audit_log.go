package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AuditLog はワークスペースの管理操作の記録です。対象が削除されても残すため外部キーは張りません
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "audit_log"}}
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("workspace_id").
			NotEmpty().
			Immutable(),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
		field.String("action").
			NotEmpty().
			Immutable(),
		field.String("target_type").
			Default("").
			Immutable(),
		field.String("target_id").
			Default("").
			Immutable(),
		// 対象が削除されても表示できるよう記録時点の名前を残す
		field.String("target_label").
			Default("").
			Immutable(),
		field.JSON("metadata", map[string]string{}).
			Immutable(),
		field.String("ip_address").
			Default("").
			Immutable(),
		field.String("user_agent").
			Default("").
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id", "created_at"),
		index.Fields("workspace_id", "actor_id"),
		index.Fields("workspace_id", "action"),
	}
}
