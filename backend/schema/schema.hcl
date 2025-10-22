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
    type = citext
  }
  column "password_hash" { type = text, null = false }
  column "display_name" { type = text, null = false }
  column "avatar_url" { type = text }
  column "created_at" { type = timestamptz, null = false, default = sql("now()") }
  column "updated_at" { type = timestamptz, null = false, default = sql("now()") }
  column "deleted_at" { type = timestamptz }
  primary_key { columns = [column.id] }
  index "users_email_idx" { columns = [column.email] }
  unique "users_email_unique" { columns = [column.email] }
}


