import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

export const workspaceRoleKeys = {
  [WorkspaceRole.UNSPECIFIED]: "member.role.member",
  [WorkspaceRole.OWNER]: "member.role.owner",
  [WorkspaceRole.ADMIN]: "member.role.admin",
  [WorkspaceRole.MEMBER]: "member.role.member",
  [WorkspaceRole.GUEST]: "member.role.guest",
} as const;
