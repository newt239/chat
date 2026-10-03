import { IconShieldCheck } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { cn } from "#/components/ui/styles/styles";
import { Switch } from "#/components/ui/Switch/Switch";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { permissions } from "#/features/admin/utils/labels";
import { tableClassNames } from "#/features/admin/utils/tableClassNames";
import { workspaceRoles } from "#/features/member/utils/workspaceRoleKeys";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

import type { Permission, PermissionGrant } from "#/gen/chat/v1/permission_service_pb";

type AdminPermissionsTabProps = {
  workspaceId: string;
  grants: readonly PermissionGrant[];
  myRole: WorkspaceRole | undefined;
};

export const AdminPermissionsTab = ({ workspaceId, grants, myRole }: AdminPermissionsTabProps) => {
  const { t } = useTranslation();
  const { updatePermission } = useAdminActions();

  // オーナーは常に許可。管理者の列は管理者が自分たちの権限を広げられないようオーナーだけが変更できる
  const isEditable = (role: WorkspaceRole) =>
    role !== WorkspaceRole.OWNER &&
    (role !== WorkspaceRole.ADMIN || myRole === WorkspaceRole.OWNER);
  const isAllowed = (role: WorkspaceRole, permission: Permission) =>
    role === WorkspaceRole.OWNER ||
    grants.some((grant) => grant.role === role && grant.permission === permission && grant.allowed);

  return (
    <div className="flex flex-col gap-3.5">
      <div className={tableClassNames.wrapper}>
        <table className={tableClassNames.table}>
          <thead>
            <tr>
              <th scope="col" className={tableClassNames.header}>
                {t("admin.permissions.operation")}
              </th>
              {workspaceRoles.map(({ key }) => (
                <th key={key} scope="col" className={cn(tableClassNames.header, "text-center")}>
                  {t(`member.role.${key}`)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {permissions.map(({ key: permissionKey, permission }) => {
              const permissionName = t(`admin.permissions.names.${permissionKey}`);
              return (
                <tr key={permission} className={tableClassNames.row}>
                  <th scope="row" className={cn(tableClassNames.cell, "text-left font-normal")}>
                    {permissionName}
                  </th>
                  {workspaceRoles.map(({ key, role }) => {
                    const label = t("admin.permissions.toggle", {
                      permission: permissionName,
                      role: t(`member.role.${key}`),
                    });
                    return (
                      <td key={key} className={tableClassNames.cell}>
                        <Switch
                          aria-label={label}
                          className="justify-center"
                          isSelected={isAllowed(role, permission)}
                          isDisabled={!isEditable(role) || updatePermission.isPending}
                          onChange={(allowed) => {
                            updatePermission.mutate(
                              { allowed, permission, role, workspaceId },
                              {
                                onSuccess: () => {
                                  toast(t("admin.permissions.updated"), { tone: "success" });
                                },
                              },
                            );
                          }}
                        >
                          {null}
                        </Switch>
                      </td>
                    );
                  })}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <p className="m-0 flex items-start gap-1.5 text-xs text-muted">
        <IconShieldCheck aria-hidden className="size-3.5 shrink-0" />
        {t("admin.permissions.note")}
      </p>
    </div>
  );
};
