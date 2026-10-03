import { AuditAction } from "#/gen/chat/v1/admin_service_pb";
import { Permission } from "#/gen/chat/v1/permission_service_pb";

// 監査ログの操作と辞書 admin.audit.actions.* のキー。絞り込みの選択肢と URL で受け付ける値もここから作る
export const auditActions = [
  { action: AuditAction.LOGIN, key: "login" },
  { action: AuditAction.LOGIN_FAILED, key: "loginFailed" },
  { action: AuditAction.MEMBER_ROLE_CHANGED, key: "memberRoleChanged" },
  { action: AuditAction.MEMBER_SUSPENDED, key: "memberSuspended" },
  { action: AuditAction.MEMBER_RESUMED, key: "memberResumed" },
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
  AuditAction.PERMISSION_CHANGED,
  AuditAction.AUDIT_LOG_EXPORTED,
]);

export const permissionKeys = {
  [Permission.UNSPECIFIED]: "unspecified",
  [Permission.CREATE_PUBLIC_CHANNEL]: "createPublicChannel",
  [Permission.CREATE_PRIVATE_CHANNEL]: "createPrivateChannel",
  [Permission.INVITE_MEMBERS]: "inviteMembers",
  [Permission.EDIT_CHANNEL_LINKS]: "editChannelLinks",
  [Permission.PIN_MESSAGES]: "pinMessages",
  [Permission.DELETE_OTHERS_MESSAGES]: "deleteOthersMessages",
  [Permission.CREATE_CUSTOM_EMOJI]: "createCustomEmoji",
} as const;

// 監査ログの詳細（metadata）はサーバーの内部名で届くため、表示名のキーに引き直す
export const permissionKeyByName: Readonly<Record<string, (typeof permissionKeys)[Permission]>> = {
  create_custom_emoji: "createCustomEmoji",
  create_private_channel: "createPrivateChannel",
  create_public_channel: "createPublicChannel",
  delete_others_messages: "deleteOthersMessages",
  edit_channel_links: "editChannelLinks",
  invite_members: "inviteMembers",
  pin_messages: "pinMessages",
};
