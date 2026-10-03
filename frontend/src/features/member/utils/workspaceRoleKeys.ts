import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

// key は辞書 member.role.* のキーで、監査ログに載るロールの内部名でもある
export const workspaceRoles = [
  { key: "owner", role: WorkspaceRole.OWNER },
  { key: "admin", role: WorkspaceRole.ADMIN },
  { key: "member", role: WorkspaceRole.MEMBER },
  { key: "guest", role: WorkspaceRole.GUEST },
] as const;

export type WorkspaceRoleKey = (typeof workspaceRoles)[number]["key"];

export const workspaceRoleKey = (role: WorkspaceRole) =>
  workspaceRoles.find((entry) => entry.role === role)?.key ?? "member";
