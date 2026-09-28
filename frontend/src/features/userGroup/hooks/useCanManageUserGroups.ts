import { useMyWorkspaceRole } from "#/features/admin/hooks/useMyWorkspaceRole";
import { isAdminRole } from "#/features/admin/utils/isAdminRole";

// ユーザーグループの作成・編集・削除・メンバーの追加削除は owner / admin だけができる（拒否は API 側でも行う）
export const useCanManageUserGroups = (workspaceId: string) =>
  isAdminRole(useMyWorkspaceRole(workspaceId).data);
