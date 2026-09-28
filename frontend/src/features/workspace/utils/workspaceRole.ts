import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

export const workspaceRoleLabels: Record<WorkspaceRole, string> = {
  [WorkspaceRole.UNSPECIFIED]: "",
  [WorkspaceRole.OWNER]: "オーナー",
  [WorkspaceRole.ADMIN]: "管理者",
  [WorkspaceRole.MEMBER]: "メンバー",
  [WorkspaceRole.GUEST]: "ゲスト",
};
