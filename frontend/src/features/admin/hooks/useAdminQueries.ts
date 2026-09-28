import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import { keepPreviousData } from "@tanstack/react-query";

import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { PermissionService } from "#/gen/chat/v1/permission_service_pb";

import type { ListAuditLogsRequestSchema } from "#/gen/chat/v1/admin_service_pb";

import type { MessageInitShape } from "@bufbuild/protobuf";

export const useAdminMembers = (workspaceId: string) =>
  useQuery(AdminService.method.listAdminMembers, { workspaceId }, { select: (res) => res.members });

export const useAuditLogs = (input: MessageInitShape<typeof ListAuditLogsRequestSchema>) =>
  useQuery(AdminService.method.listAuditLogs, input);

// 保存先を NoSQL に移せるよう総件数やオフセットは使わず、ページトークンで続きを読み込む。
// 絞り込みを変えても前の結果を表示したままにする
export const useAuditLogPages = (input: MessageInitShape<typeof ListAuditLogsRequestSchema>) =>
  useInfiniteQuery(
    AdminService.method.listAuditLogs,
    { ...input, pageToken: "" },
    {
      getNextPageParam: (res) => res.nextPageToken || undefined,
      pageParamKey: "pageToken",
      placeholderData: keepPreviousData,
    },
  );

export const usePermissions = (workspaceId: string) =>
  useQuery(PermissionService.method.getPermissions, { workspaceId });
