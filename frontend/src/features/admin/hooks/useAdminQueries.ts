import { useQuery } from "@connectrpc/connect-query";
import { keepPreviousData } from "@tanstack/react-query";

import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { PermissionService } from "#/gen/chat/v1/permission_service_pb";

import type { ListAuditLogsRequestSchema } from "#/gen/chat/v1/admin_service_pb";

import type { MessageInitShape } from "@bufbuild/protobuf";

export const useAdminMembers = (workspaceId: string) =>
  useQuery(AdminService.method.listAdminMembers, { workspaceId }, { select: (res) => res.members });

// ページや絞り込みを変えても前の結果を表示したままにする
export const useAuditLogs = (input: MessageInitShape<typeof ListAuditLogsRequestSchema>) =>
  useQuery(AdminService.method.listAuditLogs, input, { placeholderData: keepPreviousData });

export const usePermissions = (workspaceId: string) =>
  useQuery(PermissionService.method.getPermissions, { workspaceId });
