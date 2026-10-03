import { formatBytes, formatNumber } from "@chat/i18n/format";
import { useQuery } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { usePreferences } from "#/hooks/usePreferences";
import { isAdminRole } from "#/lib/isAdminRole";

import { AuditLogTable } from "./AuditLogTable";

import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

const TOP_MEMBERS = 5;

type AdminOverviewTabProps = {
  workspaceId: string;
  members: readonly AdminMember[];
};

export const AdminOverviewTab = ({ workspaceId, members }: AdminOverviewTabProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const { data: recent } = useQuery(AdminService.method.listAuditLogs, { limit: 5, workspaceId });

  const active = members.filter((member) => member.suspendedAt === undefined);
  const counts = [
    [t("admin.overview.members"), active.length],
    [t("admin.overview.admins"), active.filter((member) => isAdminRole(member.role)).length],
    [t("admin.overview.suspended"), members.length - active.length],
  ] as const;
  const rankings = [
    {
      format: (value: number) => formatNumber(value, locale),
      title: t("admin.overview.topPosters"),
      valueOf: (member: AdminMember) => member.recentMessageCount,
    },
    {
      format: (value: number) => formatBytes(value, locale),
      title: t("admin.overview.topStorage"),
      valueOf: (member: AdminMember) => Number(member.storageBytes),
    },
  ];

  return (
    <div className="flex flex-col gap-5">
      <dl className="m-0 flex flex-wrap gap-x-6 gap-y-2">
        {counts.map(([label, value]) => (
          <div key={label} className="flex items-baseline gap-2">
            <dt className="text-caption text-muted">{label}</dt>
            <dd className="m-0 text-title font-bold">{formatNumber(value, locale)}</dd>
          </div>
        ))}
      </dl>
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        {rankings.map(({ format, title, valueOf }) => (
          <section key={title} className="flex flex-col gap-2">
            <h2 className="m-0 text-body-sm font-bold">{title}</h2>
            <ol className="m-0 flex flex-col gap-1 pl-5 text-body-sm">
              {members
                .filter((member) => valueOf(member) > 0)
                .toSorted((a, b) => valueOf(b) - valueOf(a))
                .slice(0, TOP_MEMBERS)
                .map((member) => (
                  <li key={member.userId}>
                    <span className="flex justify-between gap-2">
                      <span className="truncate">{member.displayName}</span>
                      <span className="text-muted tabular-nums">{format(valueOf(member))}</span>
                    </span>
                  </li>
                ))}
            </ol>
          </section>
        ))}
      </div>
      <section className="flex flex-col gap-2">
        <header className="flex items-baseline justify-between gap-2">
          <h2 className="m-0 text-body-sm font-bold">{t("admin.overview.recentAudit")}</h2>
          <Link
            to="/app/$workspaceId/admin"
            params={{ workspaceId }}
            search={{ tab: "audit" }}
            className="text-xs"
          >
            {t("admin.overview.viewAll")}
          </Link>
        </header>
        <AuditLogTable logs={recent?.logs ?? []} />
      </section>
    </div>
  );
};
