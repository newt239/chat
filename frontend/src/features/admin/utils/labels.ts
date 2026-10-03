import { AuditAction } from "#/gen/chat/v1/admin_service_pb";
import { Permission } from "#/gen/chat/v1/permission_service_pb";

// 監査ログの操作と辞書 admin.audit.actions.* のキー。絞り込みの選択肢と URL で受け付ける値もここから作る
export const auditActions = [
  { action: AuditAction.LOGIN, key: "login" },
  { action: AuditAction.LOGIN_FAILED, key: "loginFailed" },
  { action: AuditAction.MEMBER_ROLE_CHANGED, key: "memberRoleChanged" },
  { action: AuditAction.MEMBER_SUSPENDED, key: "memberSuspended" },
  { action: AuditAction.MEMBER_RESUMED, key: "memberResumed" },
  { action: AuditAction.MEMBER_REMOVED, key: "memberRemoved" },
  { action: AuditAction.CHANNEL_CREATED, key: "channelCreated" },
  { action: AuditAction.PERMISSION_CHANGED, key: "permissionChanged" },
  { action: AuditAction.AUDIT_LOG_EXPORTED, key: "auditLogExported" },
  { action: AuditAction.CUSTOM_EMOJI_CREATED, key: "customEmojiCreated" },
  { action: AuditAction.CUSTOM_EMOJI_DELETED, key: "customEmojiDeleted" },
  { action: AuditAction.APP_CREATED, key: "appCreated" },
  { action: AuditAction.APP_DELETED, key: "appDeleted" },
] as const;

export const auditActionKey = (action: AuditAction) =>
  auditActions.find((entry) => entry.action === action)?.key ?? "unspecified";

// 権限やセキュリティに関わる操作は目立たせる
export const sensitiveAuditActions: ReadonlySet<AuditAction> = new Set([
  AuditAction.LOGIN_FAILED,
  AuditAction.MEMBER_ROLE_CHANGED,
  AuditAction.MEMBER_SUSPENDED,
  AuditAction.MEMBER_REMOVED,
  AuditAction.PERMISSION_CHANGED,
  AuditAction.AUDIT_LOG_EXPORTED,
]);

// 設定できる権限と辞書 admin.permissions.names.* のキー。name は監査ログの詳細に載るサーバーの内部名
export const permissions = [
  {
    key: "createPublicChannel",
    name: "create_public_channel",
    permission: Permission.CREATE_PUBLIC_CHANNEL,
  },
  {
    key: "createPrivateChannel",
    name: "create_private_channel",
    permission: Permission.CREATE_PRIVATE_CHANNEL,
  },
  { key: "inviteMembers", name: "invite_members", permission: Permission.INVITE_MEMBERS },
  {
    key: "editChannelLinks",
    name: "edit_channel_links",
    permission: Permission.EDIT_CHANNEL_LINKS,
  },
  { key: "pinMessages", name: "pin_messages", permission: Permission.PIN_MESSAGES },
  {
    key: "deleteOthersMessages",
    name: "delete_others_messages",
    permission: Permission.DELETE_OTHERS_MESSAGES,
  },
  {
    key: "createCustomEmoji",
    name: "create_custom_emoji",
    permission: Permission.CREATE_CUSTOM_EMOJI,
  },
] as const;
