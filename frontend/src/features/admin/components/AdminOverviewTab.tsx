import { formatBytes, formatNumber } from "@chat/i18n/format";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Link } from "#/components/ui/Link/Link";
import { useAuditLogs } from "#/features/admin/hooks/useAdminQueries";
import { ChartCard } from "#/features/insights/components/ChartCard";
import { HBarList } from "#/features/insights/components/HBarList";
import { KpiCard } from "#/features/insights/components/KpiCard";
import { usePreferences } from "#/hooks/usePreferences";
import { isAdminRole } from "#/lib/isAdminRole";

import { AuditLogTable } from "./AuditLogTable";

import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

const TOP_MEMBERS = 5;
const RECENT_LOGS = 5;

type AdminOverviewTabProps = {
  workspaceId: string;
  members: readonly AdminMember[];
};

export const AdminOverviewTab = ({ workspaceId, members }: AdminOverviewTabProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const { data: recent } = useAuditLogs({ limit: RECENT_LOGS, workspaceId });

  const active = members.filter((member) => member.suspendedAt === undefined);
  const kpis = [
    { label: t("admin.overview.members"), value: active.length },
    {
      label: t("admin.overview.admins"),
      value: active.filter((member) => isAdminRole(member.role)).length,
    },
    { label: t("admin.overview.suspended"), value: members.length - active.length },
  ];
  const ranking = (valueOf: (member: AdminMember) => number, format: (value: number) => string) =>
    members
      .filter((member) => valueOf(member) > 0)
      .toSorted((a, b) => valueOf(b) - valueOf(a))
      .slice(0, TOP_MEMBERS)
      .map((member) => ({
        icon: <Avatar name={member.displayName} src={member.avatarUrl} size={16} />,
        key: member.userId,
        label: member.displayName,
        value: valueOf(member),
        valueLabel: format(valueOf(member)),
      }));
  const posters = ranking(
    (member) => member.recentMessageCount,
    (value) => formatNumber(value, locale),
  );
  const storage = ranking(
    (member) => Number(member.storageBytes),
    (value) => formatBytes(value, locale),
  );
  const tableOf = (rows: typeof posters, column: string) => ({
    columns: [t("admin.members.columns.member"), column],
    rows: rows.map((row) => [row.label, row.valueLabel]),
  });

  return (
    <div className="flex flex-col gap-3.5">
      <div className="grid grid-cols-[repeat(auto-fit,minmax(170px,1fr))] gap-2.5 max-md:grid-cols-2 max-md:gap-2">
        {kpis.map((kpi) => (
          <KpiCard
            key={kpi.label}
            label={kpi.label}
            value={formatNumber(kpi.value, locale)}
            unit={t("insights.kpi.unitPeople")}
            delta={null}
          />
        ))}
      </div>
      <div className="grid grid-cols-1 gap-3.5 lg:grid-cols-2">
        <ChartCard
          title={t("admin.overview.topPosters")}
          note={t("admin.overview.topPostersNote")}
          table={tableOf(posters, t("admin.members.columns.messages"))}
        >
          <HBarList rows={posters} emptyLabel={t("insights.empty")} />
        </ChartCard>
        <ChartCard
          title={t("admin.overview.topStorage")}
          note={t("admin.overview.topStorageNote")}
          table={tableOf(storage, t("admin.members.columns.storage"))}
        >
          <HBarList rows={storage} emptyLabel={t("insights.empty")} />
        </ChartCard>
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
