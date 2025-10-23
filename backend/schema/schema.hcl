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
  column "password_hash" { type = text, null = false }
  column "display_name" { type = text, null = false }
  column "avatar_url" { type = text }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  column "updated_at" { type = timestamptz, null = false, default = sql("now()") }
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
  column "user_id" { type = uuid, null = false }
  column "refresh_token_hash" { type = text, null = false }
  column "expires_at" { type = timestamptz, null = false }
  column "revoked_at" { type = timestamptz }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
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
  column "name" { type = text, null = false }
  column "description" { type = text }
  column "icon_url" { type = text }
  column "created_by" { type = uuid, null = false }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  column "updated_at" { type = timestamptz, null = false, default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "workspaces_created_by_fkey" {
    columns = [column.created_by]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
}

table "workspace_members" {
  schema = schema.public
  column "workspace_id" { type = uuid, null = false }
  column "user_id" { type = uuid, null = false }
  column "role" { type = text, null = false }
  column "joined_at" { type = timestamptz, null = false, default = sql("now()") }
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
  column "workspace_id" { type = uuid, null = false }
  column "name" { type = text, null = false }
  column "description" { type = text }
  column "is_private" { type = boolean, null = false, default = false }
  column "created_by" { type = uuid, null = false }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  column "updated_at" { type = timestamptz, null = false, default = sql("now()") }
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
  column "channel_id" { type = uuid, null = false }
  column "user_id" { type = uuid, null = false }
  column "joined_at" { type = timestamptz, null = false, default = sql("now()") }
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
}

table "messages" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "channel_id" { type = uuid, null = false }
  column "user_id" { type = uuid, null = false }
  column "parent_id" { type = uuid }
  column "body" { type = text, null = false }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  column "edited_at" { type = timestamptz }
  column "deleted_at" { type = timestamptz }
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
  index "messages_channel_id_created_at_idx" { columns = [column.channel_id, column.created_at], type = BTREE }
  index "messages_parent_id_created_at_idx" { columns = [column.parent_id, column.created_at], type = BTREE }
}

table "message_reactions" {
  schema = schema.public
  column "message_id" { type = uuid, null = false }
  column "user_id" { type = uuid, null = false }
  column "emoji" { type = text, null = false }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
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
  column "channel_id" { type = uuid, null = false }
  column "user_id" { type = uuid, null = false }
  column "last_read_at" { type = timestamptz, null = false, default = sql("now()") }
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
  index "channel_read_states_user_last_read_idx" { columns = [column.user_id, column.last_read_at], type = BTREE }
}

table "attachments" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "message_id" { type = uuid, null = false }
  column "file_name" { type = text, null = false }
  column "mime_type" { type = text, null = false }
  column "size_bytes" { type = bigint, null = false }
  column "storage_key" { type = text, null = false }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "attachments_message_id_fkey" {
    columns = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_delete = CASCADE
  }
  index "attachments_message_id_idx" { columns = [column.message_id] }
}


