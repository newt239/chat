schema "public" {
}

table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    null = false
    type = text
  }
  column "password_hash" {
    type = text
    null = false
  }
  column "display_name" {
    type = text
    null = false
  }
  column "avatar_url" { 
    type = text
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  unique "users_email_unique" { columns = [column.email] }
  index "users_email_idx" { columns = [column.email] }
}

table "sessions" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "refresh_token_hash" {
    type = text
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "revoked_at" { 
    type = timestamptz
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "sessions_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "sessions_user_id_idx" { columns = [column.user_id] }
  index "sessions_expires_at_idx" { columns = [column.expires_at] }
}

table "workspaces" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "name" {
    type = text
    null = false
  }
  column "description" { 
    type = text
    null = true
  }
  column "icon_url" { 
    type = text
    null = true
  }
  column "created_by" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "workspaces_created_by_fkey" {
    columns = [column.created_by]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
}

table "workspace_members" {
  schema = schema.public
  column "workspace_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "role" {
    type = text
    null = false
  }
  column "joined_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.workspace_id, column.user_id] }
  foreign_key "workspace_members_workspace_id_fkey" {
    columns = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_delete = CASCADE
  }
  foreign_key "workspace_members_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "workspace_members_user_id_idx" { columns = [column.user_id] }
}

table "channels" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "workspace_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "description" { 
    type = text
    null = true
  }
  column "is_private" {
    type = boolean
    null = false
    default = false
  }
  column "created_by" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "channels_workspace_id_fkey" {
    columns = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_delete = CASCADE
  }
  foreign_key "channels_created_by_fkey" {
    columns = [column.created_by]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  unique "channels_workspace_name_unique" { columns = [column.workspace_id, column.name] }
  index "channels_workspace_id_idx" { columns = [column.workspace_id] }
  index "channels_workspace_id_private_idx" { columns = [column.workspace_id, column.is_private] }
}

table "channel_members" {
  schema = schema.public
  column "channel_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "role" {
    type = text
    null = false
    default = "member"
  }
  column "joined_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.channel_id, column.user_id] }
  foreign_key "channel_members_channel_id_fkey" {
    columns = [column.channel_id]
    ref_columns = [table.channels.column.id]
    on_delete = CASCADE
  }
  foreign_key "channel_members_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "channel_members_user_id_idx" { columns = [column.user_id] }
  check "channel_members_role_check" {
    expr = "role IN ('member', 'admin')"
  }
}

table "messages" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "channel_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "parent_id" { 
    type = uuid
    null = true
  }
  column "body" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "edited_at" { 
    type = timestamptz
    null = true
  }
  column "deleted_at" { 
    type = timestamptz
    null = true
  }
  column "deleted_by" { 
    type = uuid
    null = true
  }
  primary_key { columns = [column.id] }
  foreign_key "messages_channel_id_fkey" {
    columns = [column.channel_id]
    ref_columns = [table.channels.column.id]
    on_delete = CASCADE
  }
  foreign_key "messages_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "messages_parent_id_fkey" {
    columns = [column.parent_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  index "messages_channel_id_created_at_idx" {
    columns = [column.channel_id, column.created_at]
    type = BTREE
  }
  index "messages_parent_id_created_at_idx" {
    columns = [column.parent_id, column.created_at]
    type = BTREE
  }
}

table "message_reactions" {
  schema = schema.public
  column "message_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "emoji" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.message_id, column.user_id, column.emoji] }
  foreign_key "message_reactions_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  foreign_key "message_reactions_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
}

table "channel_read_states" {
  schema = schema.public
  column "channel_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "last_read_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.channel_id, column.user_id] }
  foreign_key "channel_read_states_channel_id_fkey" {
    columns = [column.channel_id]
    ref_columns = [table.channels.column.id]
    on_delete = CASCADE
  }
  foreign_key "channel_read_states_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "channel_read_states_user_last_read_idx" {
    columns = [column.user_id]
    type = BTREE
  }
}

table "attachments" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "message_id" { 
    type = uuid
    null = true
  }
  column "uploader_id" {
    type = uuid
    null = false
  }
  column "channel_id" {
    type = uuid
    null = false
  }
  column "file_name" {
    type = text
    null = false
  }
  column "mime_type" {
    type = text
    null = false
  }
  column "size_bytes" {
    type = bigint
    null = false
  }
  column "storage_key" {
    type = text
    null = false
  }
  column "status" {
    type = text
    null = false
    default = "pending"
  }
  column "uploaded_at" { 
    type = timestamptz
    null = true
  }
  column "expires_at" { 
    type = timestamptz
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "attachments_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  foreign_key "attachments_uploader_id_fkey" {
    columns = [column.uploader_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "attachments_channel_id_fkey" {
    columns = [column.channel_id]
    ref_columns = [table.channels.column.id]
    on_delete = CASCADE
  }
  index "attachments_message_id_idx" { columns = [column.message_id] }
  index "attachments_uploader_status_idx" { columns = [column.uploader_id, column.status] }
  index "attachments_channel_id_idx" { columns = [column.channel_id] }
}

table "user_groups" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "workspace_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "description" { 
    type = text
    null = true
  }
  column "created_by" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "user_groups_workspace_id_fkey" {
    columns = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_delete = CASCADE
  }
  foreign_key "user_groups_created_by_fkey" {
    columns = [column.created_by]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  unique "user_groups_workspace_name_unique" { columns = [column.workspace_id, column.name] }
  index "user_groups_workspace_id_idx" { columns = [column.workspace_id] }
}

table "user_group_members" {
  schema = schema.public
  column "group_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "joined_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.group_id, column.user_id] }
  foreign_key "user_group_members_group_id_fkey" {
    columns = [column.group_id]
    ref_columns = [table.user_groups.column.id]
    on_delete = CASCADE
  }
  foreign_key "user_group_members_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "user_group_members_user_id_idx" { columns = [column.user_id] }
}

table "message_user_mentions" {
  schema = schema.public
  column "message_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.message_id, column.user_id] }
  foreign_key "message_user_mentions_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  foreign_key "message_user_mentions_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "message_user_mentions_user_id_idx" { columns = [column.user_id] }
}

table "message_group_mentions" {
  schema = schema.public
  column "message_id" {
    type = uuid
    null = false
  }
  column "group_id" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.message_id, column.group_id] }
  foreign_key "message_group_mentions_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  foreign_key "message_group_mentions_group_id_fkey" {
    columns = [column.group_id]
    ref_columns = [table.user_groups.column.id]
    on_delete = CASCADE
  }
  index "message_group_mentions_group_id_idx" { columns = [column.group_id] }
}

table "message_links" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "message_id" {
    type = uuid
    null = false
  }
  column "url" {
    type = text
    null = false
  }
  column "title" { 
    type = text
    null = true
  }
  column "description" { 
    type = text
    null = true
  }
  column "image_url" { 
    type = text
    null = true
  }
  column "site_name" { 
    type = text
    null = true
  }
  column "card_type" { 
    type = text
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "message_links_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  unique "message_links_url_unique" { columns = [column.url] }
  index "message_links_message_id_idx" { columns = [column.message_id] }
}

table "message_bookmarks" {
  schema = schema.public
  column "user_id" {
    type = uuid
    null = false
  }
  column "message_id" {
    type = uuid
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.user_id, column.message_id] }
  foreign_key "message_bookmarks_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "message_bookmarks_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  index "message_bookmarks_user_id_idx" { columns = [column.user_id] }
  index "message_bookmarks_created_at_idx" { columns = [column.created_at] }
}

table "thread_metadata" {
  schema = schema.public
  column "message_id" {
    type = uuid
    null = false
  }
  column "reply_count" {
    type = integer
    null = false
    default = 0
  }
  column "last_reply_at" { 
    type = timestamptz
    null = true
  }
  column "last_reply_user_id" { 
    type = uuid
    null = true
  }
  column "participant_user_ids" {
    type = sql("uuid[]")
    null = false
    default = sql("ARRAY[]::uuid[]")
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key { columns = [column.message_id] }
  foreign_key "thread_metadata_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  foreign_key "thread_metadata_last_reply_user_id_fkey" {
    columns = [column.last_reply_user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
  }
  index "thread_metadata_last_reply_at_idx" {
    columns = [column.last_reply_at]
    type = BTREE
  }
}


