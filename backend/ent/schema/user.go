package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user"}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password_hash").
			NotEmpty(),
		// 初回の Google ログインで紐付ける
		field.String("google_sub").
			Optional().
			Nillable().
			Unique(),
		field.String("display_name").
			NotEmpty(),
		field.String("bio").
			Optional().
			Nillable(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		// アプリの投稿名義。ログインできず、ワークスペースのメンバーにもならない
		field.Bool("is_app").
			Default(false),
		// 公式アプリの投稿名義。この名義の投稿は誰も削除・編集できない
		field.Bool("is_official").
			Default(false),
		// 退会済み。投稿の名義として行だけ残し、個人情報は匿名化する
		field.Time("deleted_at").
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

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel_members", ChannelMember.Type).
			Ref("user"),
		edge.To("preference", UserPreference.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("links", UserLink.Type).
			Ref("user"),
	}
}
