import { createQueryOptions } from "@connectrpc/connect-query";
import { createFileRoute, redirect } from "@tanstack/react-router";

import { AdminPage } from "#/features/admin/components/AdminPage";
import { adminSearchSchema } from "#/features/admin/schemas";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { transport } from "#/lib/api/transport";
import { isAdminRole } from "#/lib/isAdminRole";
import { queryClient } from "#/providers/query/query";

// 一般メンバーは管理画面に入れずワークスペースのトップへ戻す（API 側でも拒否される）
export const Route = createFileRoute("/app/$workspaceId/admin")({
  beforeLoad: async ({ params }) => {
    const { workspace } = await queryClient.query(
      createQueryOptions(
        WorkspaceService.method.getWorkspace,
        { workspaceId: params.workspaceId },
        { transport },
      ),
    );
    if (!isAdminRole(workspace?.role)) {
      throw redirect({ params, to: "/app/$workspaceId" });
    }
  },
  component: AdminPage,
  validateSearch: adminSearchSchema,
});
