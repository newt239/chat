package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserPreference は端末をまたいで共有する表示・通知の設定です。変更したときに作るため、ない場合は既定値を使う
type UserPreference struct {
	ent.Schema
}

func (UserPreference) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_preference"}}
}

func (UserPreference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Unique().
			Immutable(),
		field.Int("theme_hue"),
		field.Float("theme_chroma"),
		field.Enum("theme_sidebar").
			Values("tinted", "light"),
		field.Enum("color_mode").
			Values("light", "dark", "system"),
		field.String("locale"),
		field.Enum("notification_level").
			Values("all", "mentions", "none"),
		// IANA のタイムゾーン名。空は未設定
		field.String("timezone"),
		// 端末のタイムゾーンが変わったら尋ねずに更新する
		field.Bool("timezone_auto_update"),
		field.Enum("channel_sort_order").
			Values("default", "recent_activity"),
		// チャンネルへの参加・追加のシステムメッセージを隠す
		field.Bool("hide_join_messages"),
	}
}

func (UserPreference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("preference").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
	}
}
