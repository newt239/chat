package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Message struct {
	ent.Schema
}

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		// 外部キーをフィールドとして公開し、ID を得るためだけに関連を読み込まずに済ませる（列名は従来のまま）
		field.UUID("channel_id", uuid.UUID{}).
			StorageKey("message_channel").
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			StorageKey("message_user").
			Immutable(),
		field.UUID("parent_id", uuid.UUID{}).
			StorageKey("message_parent").
			Optional().
			Nillable().
			Immutable(),
		// 添付や位置情報だけのメッセージは本文が空になる
		field.Text("body"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("edited_at").
			Optional(),
		field.Time("deleted_at").
			Optional(),
		field.UUID("deleted_by", uuid.UUID{}).
			Optional(),
		// Webhook が投稿ごとに指定した表示名とアイコン
		field.String("sender_name").
			Optional().
			Nillable(),
		field.String("sender_avatar_url").
			Optional().
			Nillable(),
		// 共有された位置情報。緯度と経度は両方そろって設定される
		field.Float("location_latitude").
			Optional().
			Nillable(),
		field.Float("location_longitude").
			Optional().
			Nillable(),
		field.Float("location_accuracy").
			Optional().
			Nillable(),
		field.String("location_label").
			Optional().
			Nillable(),
		// 本文の <@channel> / <@here>。届く範囲は読むときのチャンネルメンバーで決まる
		field.Bool("mentions_channel").
			Default(false),
		field.Bool("mentions_here").
			Default(false),
	}
}

func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Unique().
			Required().
			Immutable(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Immutable(),
		edge.To("parent", Message.Type).
			Field("parent_id").
			Unique().
			Immutable(),
		edge.From("replies", Message.Type).
			Ref("parent"),
		edge.From("reactions", MessageReaction.Type).
			Ref("message"),
		edge.From("bookmarks", MessageBookmark.Type).
			Ref("message"),
		edge.From("user_mentions", MessageUserMention.Type).
			Ref("message"),
		edge.From("group_mentions", MessageGroupMention.Type).
			Ref("message"),
		edge.From("links", MessageLink.Type).
			Ref("message"),
		edge.From("attachments", Attachment.Type).
			Ref("message"),
		edge.From("pins", MessagePin.Type).
			Ref("message"),
		edge.From("user_thread_follows", UserThreadFollow.Type).
			Ref("thread"),
		edge.From("thread_read_states", ThreadReadState.Type).
			Ref("thread"),
	}
}

func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		// タイムラインと未読数は削除済みを読まないため部分インデックスにする
		index.Fields("channel_id", "created_at").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("parent_id", "created_at"),
		index.Fields("user_id"),
	}
}
