import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

/** ワークスペースメンバーの削除・ロール変更。メンバー一覧と管理画面（監査ログを含む）を取り直す */
export const useWorkspaceMemberActions = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const options = {
    onError: (error: Error) => {
      toast(t("workspace.members.actionFailed"), { description: error.message, tone: "danger" });
    },
    onSuccess: (_: object, { workspaceId }: { workspaceId?: string }) =>
      Promise.all([
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            cardinality: "finite",
            input: { workspaceId },
            schema: WorkspaceService.method.listMembers,
          }),
        }),
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({ cardinality: undefined, schema: AdminService }),
        }),
      ]),
  };

  return {
    remove: useMutation(WorkspaceService.method.removeMember, options),
    updateRole: useMutation(WorkspaceService.method.updateMemberRole, options),
  };
};
