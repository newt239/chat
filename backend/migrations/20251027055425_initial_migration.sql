-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "display_name" text NOT NULL,
  "avatar_url" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_unique" UNIQUE ("email")
);
-- Create index "users_email_idx" to table: "users"
CREATE INDEX "users_email_idx" ON "public"."users" ("email");
-- Create "workspaces" table
CREATE TABLE "public"."workspaces" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NULL,
  "icon_url" text NULL,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "workspaces_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "channels" table
CREATE TABLE "public"."channels" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "workspace_id" uuid NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "is_private" boolean NOT NULL DEFAULT false,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "channels_workspace_name_unique" UNIQUE ("workspace_id", "name"),
  CONSTRAINT "channels_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "channels_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "channels_workspace_id_idx" to table: "channels"
CREATE INDEX "channels_workspace_id_idx" ON "public"."channels" ("workspace_id");
-- Create index "channels_workspace_id_private_idx" to table: "channels"
CREATE INDEX "channels_workspace_id_private_idx" ON "public"."channels" ("workspace_id", "is_private");
-- Create "messages" table
CREATE TABLE "public"."messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "channel_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "parent_id" uuid NULL,
  "body" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "edited_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messages_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "public"."channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "messages_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "messages_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "messages_channel_id_created_at_idx" to table: "messages"
CREATE INDEX "messages_channel_id_created_at_idx" ON "public"."messages" ("channel_id", "created_at");
-- Create index "messages_parent_id_created_at_idx" to table: "messages"
CREATE INDEX "messages_parent_id_created_at_idx" ON "public"."messages" ("parent_id", "created_at");
-- Create "attachments" table
CREATE TABLE "public"."attachments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "message_id" uuid NULL,
  "uploader_id" uuid NOT NULL,
  "channel_id" uuid NOT NULL,
  "file_name" text NOT NULL,
  "mime_type" text NOT NULL,
  "size_bytes" bigint NOT NULL,
  "storage_key" text NOT NULL,
  "status" text NOT NULL DEFAULT 'pending',
  "uploaded_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "attachments_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "public"."channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "attachments_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "attachments_uploader_id_fkey" FOREIGN KEY ("uploader_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "attachments_channel_id_idx" to table: "attachments"
CREATE INDEX "attachments_channel_id_idx" ON "public"."attachments" ("channel_id");
-- Create index "attachments_message_id_idx" to table: "attachments"
CREATE INDEX "attachments_message_id_idx" ON "public"."attachments" ("message_id");
-- Create index "attachments_uploader_status_idx" to table: "attachments"
CREATE INDEX "attachments_uploader_status_idx" ON "public"."attachments" ("uploader_id", "status");
-- Create "channel_members" table
CREATE TABLE "public"."channel_members" (
  "channel_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" text NOT NULL DEFAULT 'member',
  "joined_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("channel_id", "user_id"),
  CONSTRAINT "channel_members_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "public"."channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "channel_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "channel_members_role_check" CHECK (role = ANY (ARRAY['member'::text, 'admin'::text]))
);
-- Create index "channel_members_user_id_idx" to table: "channel_members"
CREATE INDEX "channel_members_user_id_idx" ON "public"."channel_members" ("user_id");
-- Create "channel_read_states" table
CREATE TABLE "public"."channel_read_states" (
  "channel_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "last_read_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("channel_id", "user_id"),
  CONSTRAINT "channel_read_states_channel_id_fkey" FOREIGN KEY ("channel_id") REFERENCES "public"."channels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "channel_read_states_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "channel_read_states_user_last_read_idx" to table: "channel_read_states"
CREATE INDEX "channel_read_states_user_last_read_idx" ON "public"."channel_read_states" ("user_id");
-- Create "message_bookmarks" table
CREATE TABLE "public"."message_bookmarks" (
  "user_id" uuid NOT NULL,
  "message_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id", "message_id"),
  CONSTRAINT "message_bookmarks_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "message_bookmarks_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "message_bookmarks_created_at_idx" to table: "message_bookmarks"
CREATE INDEX "message_bookmarks_created_at_idx" ON "public"."message_bookmarks" ("created_at");
-- Create index "message_bookmarks_user_id_idx" to table: "message_bookmarks"
CREATE INDEX "message_bookmarks_user_id_idx" ON "public"."message_bookmarks" ("user_id");
-- Create "user_groups" table
CREATE TABLE "public"."user_groups" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "workspace_id" uuid NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "created_by" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_groups_workspace_name_unique" UNIQUE ("workspace_id", "name"),
  CONSTRAINT "user_groups_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_groups_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "user_groups_workspace_id_idx" to table: "user_groups"
CREATE INDEX "user_groups_workspace_id_idx" ON "public"."user_groups" ("workspace_id");
-- Create "message_group_mentions" table
CREATE TABLE "public"."message_group_mentions" (
  "message_id" uuid NOT NULL,
  "group_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("message_id", "group_id"),
  CONSTRAINT "message_group_mentions_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "public"."user_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "message_group_mentions_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "message_group_mentions_group_id_idx" to table: "message_group_mentions"
CREATE INDEX "message_group_mentions_group_id_idx" ON "public"."message_group_mentions" ("group_id");
-- Create "message_links" table
CREATE TABLE "public"."message_links" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "message_id" uuid NOT NULL,
  "url" text NOT NULL,
  "title" text NULL,
  "description" text NULL,
  "image_url" text NULL,
  "site_name" text NULL,
  "card_type" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "message_links_url_unique" UNIQUE ("url"),
  CONSTRAINT "message_links_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "message_links_message_id_idx" to table: "message_links"
CREATE INDEX "message_links_message_id_idx" ON "public"."message_links" ("message_id");
-- Create "message_reactions" table
CREATE TABLE "public"."message_reactions" (
  "message_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "emoji" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("message_id", "user_id", "emoji"),
  CONSTRAINT "message_reactions_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "message_reactions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "message_user_mentions" table
CREATE TABLE "public"."message_user_mentions" (
  "message_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("message_id", "user_id"),
  CONSTRAINT "message_user_mentions_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "message_user_mentions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "message_user_mentions_user_id_idx" to table: "message_user_mentions"
CREATE INDEX "message_user_mentions_user_id_idx" ON "public"."message_user_mentions" ("user_id");
-- Create "sessions" table
CREATE TABLE "public"."sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "refresh_token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sessions_expires_at_idx" to table: "sessions"
CREATE INDEX "sessions_expires_at_idx" ON "public"."sessions" ("expires_at");
-- Create index "sessions_user_id_idx" to table: "sessions"
CREATE INDEX "sessions_user_id_idx" ON "public"."sessions" ("user_id");
-- Create "thread_metadata" table
CREATE TABLE "public"."thread_metadata" (
  "message_id" uuid NOT NULL,
  "reply_count" integer NOT NULL DEFAULT 0,
  "last_reply_at" timestamptz NULL,
  "last_reply_user_id" uuid NULL,
  "participant_user_ids" uuid[] NOT NULL DEFAULT ARRAY[]::uuid[],
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("message_id"),
  CONSTRAINT "thread_metadata_last_reply_user_id_fkey" FOREIGN KEY ("last_reply_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "thread_metadata_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "thread_metadata_last_reply_at_idx" to table: "thread_metadata"
CREATE INDEX "thread_metadata_last_reply_at_idx" ON "public"."thread_metadata" ("last_reply_at");
-- Create "user_group_members" table
CREATE TABLE "public"."user_group_members" (
  "group_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "joined_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("group_id", "user_id"),
  CONSTRAINT "user_group_members_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "public"."user_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_group_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "user_group_members_user_id_idx" to table: "user_group_members"
CREATE INDEX "user_group_members_user_id_idx" ON "public"."user_group_members" ("user_id");
-- Create "workspace_members" table
CREATE TABLE "public"."workspace_members" (
  "workspace_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" text NOT NULL,
  "joined_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "user_id"),
  CONSTRAINT "workspace_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "workspace_members_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "workspace_members_user_id_idx" to table: "workspace_members"
CREATE INDEX "workspace_members_user_id_idx" ON "public"."workspace_members" ("user_id");
