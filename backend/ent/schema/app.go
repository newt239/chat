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

// App はワークスペースに追加する連携アプリです。投稿は専用のボットユーザー名義で行い、できることは permissions で決まる
type App struct {
	ent.Schema
}

func (App) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "app"}}
}

func (App) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("created_by_id", uuid.UUID{}),
		field.UUID("bot_user_id", uuid.UUID{}),
		field.UUID("default_channel_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("workspace_id").
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		// 着信 Webhook のトークンのハッシュ。公式アプリは外から投稿できないため持たない
		field.String("token_hash").
			Optional().
			Nillable().
			Sensitive(),
		field.Strings("permissions").
			Default([]string{}),
		// 参加しているチャンネルでの投稿を送る先と、署名に使う秘密鍵
		field.String("outgoing_url").
			Optional().
			Nillable(),
		field.String("outgoing_secret").
			Optional().
			Nillable().
			Sensitive(),
		// ワークスペースごとに 1 つ用意する公式アプリ。編集・削除できず、投稿も削除できない
		field.Bool("is_official").
			Default(false),
		field.Time("last_used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (App) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("created_by", User.Type).
			Field("created_by_id").
			Unique().
			Required(),
		// 削除後も過去の投稿の名義として残すため、ボットユーザーは消さない
		edge.To("bot_user", User.Type).
			Field("bot_user_id").
			Unique().
			Required(),
		// 着信 Webhook で投稿先を省略したときのチャンネル
		edge.To("default_channel", Channel.Type).
			Field("default_channel_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (App) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id"),
		// 公式アプリはワークスペースに 1 つだけにする
		index.Fields("workspace_id").
			Unique().
			StorageKey("app_official_workspace_id").
			Annotations(entsql.IndexWhere("is_official")),
	}
}
