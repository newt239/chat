import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { PermissionService } from "#/gen/chat/v1/permission_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";

// 操作は監査ログにも残るため、管理画面の問い合わせはまとめて取り直す
const affectedServices = [AdminService, PermissionService, WorkspaceService];

export const useAdminActions = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const options = {
    onError: (error: Error) => {
      toast(t("admin.members.actionFailed"), { description: error.message, tone: "danger" });
    },
    onSuccess: () =>
      Promise.all(
        affectedServices.map((schema) =>
          queryClient.invalidateQueries({
            queryKey: createConnectQueryKey({ cardinality: undefined, schema }),
          }),
        ),
      ),
  };

  return {
    exportAuditLogs: useMutation(AdminService.method.exportAuditLogs, options),
    resume: useMutation(AdminService.method.resumeMember, options),
    suspend: useMutation(AdminService.method.suspendMember, options),
    updatePermission: useMutation(PermissionService.method.updatePermission, options),
    updateRole: useMutation(WorkspaceService.method.updateMemberRole, options),
    updateWorkspace: useMutation(WorkspaceService.method.updateWorkspace, options),
  };
};
