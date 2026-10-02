import { useState } from "react";

import { formatBytes, formatNumber } from "@chat/i18n";
import { IconSearch } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { Input, SearchField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Select } from "#/components/ui/Select/Select";
import { cn } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { tableClassNames } from "#/features/admin/utils/tableClassNames";
import { summarizeUserAgent } from "#/features/admin/utils/userAgent";
import { workspaceRoleKeys } from "#/features/member/utils/workspaceRoleKeys";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";
import { userAtom } from "#/providers/store/auth";

import { MemberSuspendButton } from "./MemberSuspendButton";
import { RoleSelect } from "./RoleSelect";

import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

const roleFilters = {
  admin: WorkspaceRole.ADMIN,
  all: null,
  guest: WorkspaceRole.GUEST,
  member: WorkspaceRole.MEMBER,
  owner: WorkspaceRole.OWNER,
} as const;
type RoleFilter = keyof typeof roleFilters;
const roleFilterValues = ["all", "owner", "admin", "member", "guest"] as const;

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
  const { formatDateTime, formatRelativeTime, locale } = useDateFormat();
  const myId = useAtomValue(userAtom)?.id;
  const { updateRole } = useAdminActions();
  const [query, setQuery] = useState("");
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("all");
  const now = new Date();

  const normalizedQuery = query.trim().toLowerCase();
  const rows = members.filter(
    (member) =>
      (roleFilters[roleFilter] === null || member.role === roleFilters[roleFilter]) &&
      `${member.displayName} ${member.email}`.toLowerCase().includes(normalizedQuery),
  );
  const roleLabel = (role: WorkspaceRole) => t(workspaceRoleKeys[role]);

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-wrap items-end gap-2">
        <SearchField
          value={query}
          onChange={setQuery}
          aria-label={t("admin.members.search")}
          className="flex h-[34px] min-w-0 flex-[1_1_200px] items-center gap-2 rounded-md border border-border-strong bg-surface px-2.5 text-muted data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft"
        >
          <IconSearch aria-hidden className="size-[15px] shrink-0" />
          <Input
            placeholder={t("admin.members.search")}
            className="h-full min-w-0 flex-1 border-0 bg-transparent font-sans text-[13.5px] text-text outline-none placeholder:text-subtle [&::-webkit-search-cancel-button]:hidden"
          />
        </SearchField>
        <Select
          label={t("admin.members.columns.role")}
          className="w-40"
          value={roleFilter}
          onChange={setRoleFilter}
          options={roleFilterValues.map((value) => ({
            label: value === "all" ? t("admin.members.allRoles") : t(`member.role.${value}`),
            value,
          }))}
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
                        <small className="text-[11.5px] text-muted">{member.email}</small>
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
                    <span className="inline-flex items-center gap-[5px] text-[12.5px]">
                      <i
                        aria-hidden
                        className={cn(
                          "block size-[7px] rounded-full",
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
                        {formatRelativeTime(lastLogin, now)}
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
                      <MemberSuspendButton workspaceId={workspaceId} member={member} />
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <p className="m-0 text-xs text-muted">{t("admin.members.count", { count: rows.length })}</p>
    </div>
  );
};
