import { useAtomValue } from "jotai";

import { useMembers } from "#/features/member/hooks/useMembers";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { userAtom } from "#/providers/store/auth";

/** 自分がワークスペースのオーナーか管理者か */
export const useIsWorkspaceAdmin = (workspaceId: string) => {
  const userId = useAtomValue(userAtom)?.id;
  const { data: members } = useMembers(workspaceId);
  const role = members?.find((member) => member.userId === userId)?.role;
  return role === WorkspaceRole.OWNER || role === WorkspaceRole.ADMIN;
};
