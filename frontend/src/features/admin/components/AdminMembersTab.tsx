import { useState } from "react";

import { formatBytes, formatNumber, formatRelativeTime } from "@chat/i18n/format";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { SearchField } from "#/components/ui/SearchField/SearchField";
import { Select } from "#/components/ui/Select/Select";
import { cn } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { tableClassNames } from "#/features/admin/utils/tableClassNames";
import { summarizeUserAgent } from "#/features/admin/utils/userAgent";
import { workspaceRoleKey, workspaceRoles } from "#/features/member/utils/workspaceRoleKeys";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";
import { myUserIdAtom } from "#/providers/store/auth";

import { MemberSuspendButton } from "./MemberSuspendButton";
import { RoleSelect } from "./RoleSelect";

import type { WorkspaceRoleKey } from "#/features/member/utils/workspaceRoleKeys";
import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

type RoleFilter = WorkspaceRoleKey | "all";

const memberColumns = [
  { isNumeric: false, key: "member" },
  { isNumeric: false, key: "role" },
  { isNumeric: false, key: "status" },
  { isNumeric: false, key: "lastLogin" },
  { isNumeric: false, key: "device" },
  { isNumeric: true, key: "messages" },
  { isNumeric: true, key: "storage" },
] as const;

type AdminMembersTabProps = {
  workspaceId: string;
  members: readonly AdminMember[];
};

export const AdminMembersTab = ({ workspaceId, members }: AdminMembersTabProps) => {
  const { t } = useTranslation();
  const { formatDateTime, locale } = useDateFormat();
  const myId = useAtomValue(myUserIdAtom);
  const { remove, updateRole } = useAdminActions();
  const [query, setQuery] = useState("");
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("all");
  const [removing, setRemoving] = useState<AdminMember | null>(null);
  const now = new Date();

  const normalizedQuery = query.trim().toLowerCase();
  const rows = members.filter(
    (member) =>
      (roleFilter === "all" || workspaceRoleKey(member.role) === roleFilter) &&
      `${member.displayName} ${member.email}`.toLowerCase().includes(normalizedQuery),
  );
  const roleLabel = (role: WorkspaceRole) => t(`member.role.${workspaceRoleKey(role)}`);

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-wrap items-end gap-2">
        <SearchField
          value={query}
          onChange={setQuery}
          label={t("admin.members.search")}
          className="flex-1 basis-50"
        />
        <Select
          label={t("admin.members.columns.role")}
          className="w-40"
          value={roleFilter}
          onChange={setRoleFilter}
          options={[
            { label: t("admin.members.allRoles"), value: "all" },
            ...workspaceRoles.map(({ key }) => ({ label: t(`member.role.${key}`), value: key })),
          ]}
        />
      </div>
      <div className={tableClassNames.wrapper}>
        <table className={tableClassNames.table}>
          <thead>
            <tr>
              {memberColumns.map(({ key, isNumeric }) => (
                <th
                  key={key}
                  scope="col"
                  className={cn(tableClassNames.header, isNumeric && "text-right")}
                >
                  {t(`admin.members.columns.${key}`)}
                </th>
              ))}
              <th scope="col" className={tableClassNames.header}>
                <span className="sr-only">{t("admin.members.columns.actions")}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map((member) => {
              const isMe = member.userId === myId;
              const isOwner = member.role === WorkspaceRole.OWNER;
              const isSuspended = member.suspendedAt !== undefined;
              const lastLogin =
                member.lastLoginAt === undefined ? null : toDate(member.lastLoginAt);
              return (
                <tr key={member.userId} className={tableClassNames.row}>
                  <td className={tableClassNames.cell}>
                    <div className="flex items-center gap-2.5">
                      <Avatar name={member.displayName} src={member.avatarUrl} size={28} />
                      <div className="flex min-w-0 flex-col">
                        <span>
                          {member.displayName}
                          {isMe && t("admin.members.you")}
                        </span>
                        <small className="text-caption text-muted">{member.email}</small>
                      </div>
                    </div>
                  </td>
                  <td className={tableClassNames.cell}>
                    {isOwner ? (
                      roleLabel(member.role)
                    ) : (
                      <RoleSelect
                        ariaLabel={t("admin.members.roleLabel", { name: member.displayName })}
                        value={member.role}
                        isDisabled={isMe || isSuspended}
                        onChange={(role) => {
                          updateRole.mutate(
                            { role, userId: member.userId, workspaceId },
                            {
                              onSuccess: () => {
                                toast(
                                  t("admin.members.roleChanged", {
                                    name: member.displayName,
                                    role: roleLabel(role),
                                  }),
                                  { tone: "success" },
                                );
                              },
                            },
                          );
                        }}
                      />
                    )}
                  </td>
                  <td className={tableClassNames.cell}>
                    <span className="inline-flex items-center gap-1.25 text-label font-normal">
                      <i
                        aria-hidden
                        className={cn(
                          "block size-1.75 rounded-full",
                          isSuspended ? "bg-danger" : "bg-success",
                        )}
                      />
                      {isSuspended
                        ? t("admin.members.status.suspended")
                        : t("admin.members.status.active")}
                    </span>
                  </td>
                  <td className={cn(tableClassNames.cell, tableClassNames.numeric)}>
                    {lastLogin === null ? (
                      <span className="text-subtle">{t("admin.members.never")}</span>
                    ) : (
                      <time dateTime={lastLogin.toISOString()} title={formatDateTime(lastLogin)}>
                        {formatRelativeTime(lastLogin, now, locale)}
                      </time>
                    )}
                  </td>
                  <td className={cn(tableClassNames.cell, tableClassNames.numeric, "text-muted")}>
                    {lastLogin === null
                      ? "—"
                      : `${member.lastLoginIp || "—"} · ${summarizeUserAgent(member.lastLoginUserAgent) ?? t("admin.members.unknownDevice")}`}
                  </td>
                  <td className={cn(tableClassNames.cell, tableClassNames.numeric, "text-right")}>
                    {formatNumber(member.recentMessageCount, locale)}
                  </td>
                  <td className={cn(tableClassNames.cell, tableClassNames.numeric, "text-right")}>
                    {formatBytes(Number(member.storageBytes), locale)}
                  </td>
                  <td className={cn(tableClassNames.cell, "text-right")}>
                    {!isMe && !isOwner && (
                      <div className="flex justify-end gap-1.5">
                        <MemberSuspendButton workspaceId={workspaceId} member={member} />
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label={t("admin.members.removeLabel", { name: member.displayName })}
                          onPress={() => {
                            setRemoving(member);
                          }}
                        >
                          {t("admin.members.remove")}
                        </Button>
                      </div>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <p className="m-0 text-xs text-muted">{t("admin.members.count", { count: rows.length })}</p>
      <AlertDialog
        isOpen={removing !== null}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            setRemoving(null);
          }
        }}
        title={t("admin.members.removeTitle", { name: removing?.displayName ?? "" })}
        confirmLabel={t("admin.members.removeConfirm")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          if (removing !== null) {
            remove.mutate(
              { userId: removing.userId, workspaceId },
              {
                onSettled: () => {
                  setRemoving(null);
                },
              },
            );
          }
        }}
      >
        <p className="m-0">{t("admin.members.removeBody")}</p>
      </AlertDialog>
    </div>
  );
};
