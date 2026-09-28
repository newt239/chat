import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

export const workspaceRoleLabels: Record<WorkspaceRole, string> = {
  [WorkspaceRole.UNSPECIFIED]: "",
  [WorkspaceRole.OWNER]: "オーナー",
  [WorkspaceRole.ADMIN]: "管理者",
  [WorkspaceRole.MEMBER]: "メンバー",
  [WorkspaceRole.GUEST]: "ゲスト",
};

// 辞書のキー（workspace.role.*）に使う名前
export const workspaceRoleNames = {
  [WorkspaceRole.UNSPECIFIED]: "member",
  [WorkspaceRole.OWNER]: "owner",
  [WorkspaceRole.ADMIN]: "admin",
  [WorkspaceRole.MEMBER]: "member",
  [WorkspaceRole.GUEST]: "guest",
} as const;
