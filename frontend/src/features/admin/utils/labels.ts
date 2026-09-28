import { AuditAction } from "#/gen/chat/v1/admin_service_pb";
import { Permission } from "#/gen/chat/v1/permission_service_pb";

export const auditActionKeys = {
  [AuditAction.UNSPECIFIED]: "unspecified",
  [AuditAction.LOGIN]: "login",
  [AuditAction.LOGIN_FAILED]: "loginFailed",
  [AuditAction.MEMBER_ROLE_CHANGED]: "memberRoleChanged",
  [AuditAction.MEMBER_SUSPENDED]: "memberSuspended",
  [AuditAction.MEMBER_RESUMED]: "memberResumed",
  [AuditAction.CHANNEL_CREATED]: "channelCreated",
  [AuditAction.CHANNEL_DELETED]: "channelDeleted",
  [AuditAction.CHANNEL_ARCHIVED]: "channelArchived",
  [AuditAction.CHANNEL_UNARCHIVED]: "channelUnarchived",
  [AuditAction.PERMISSION_CHANGED]: "permissionChanged",
  [AuditAction.AUDIT_LOG_EXPORTED]: "auditLogExported",
} as const;

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
} as const;

// 監査ログの詳細（metadata）はサーバーの内部名で届くため、表示名のキーに引き直す
export const permissionKeyByName: Readonly<Record<string, (typeof permissionKeys)[Permission]>> = {
  create_private_channel: "createPrivateChannel",
  create_public_channel: "createPublicChannel",
  delete_others_messages: "deleteOthersMessages",
  edit_channel_links: "editChannelLinks",
  invite_members: "inviteMembers",
  pin_messages: "pinMessages",
};

export const roleKeyByName: Readonly<Record<string, "owner" | "admin" | "member" | "guest">> = {
  admin: "admin",
  guest: "guest",
  member: "member",
  owner: "owner",
};
