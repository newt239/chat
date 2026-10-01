import { useTranslation } from "react-i18next";

import { cn } from "#/components/ui/styles/styles";
import { useAuditLogFormatter } from "#/features/admin/hooks/useAuditLogFormatter";
import { auditActionKeys, sensitiveAuditActions } from "#/features/admin/utils/labels";
import { tableClassNames } from "#/features/admin/utils/tableClassNames";
import { summarizeUserAgent } from "#/features/admin/utils/userAgent";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";

import type { AuditLog } from "#/gen/chat/v1/admin_service_pb";

const columns = ["time", "actor", "action", "target", "detail", "source"] as const;

type AuditLogTableProps = {
  logs: readonly AuditLog[];
};

export const AuditLogTable = ({ logs }: AuditLogTableProps) => {
  const { t } = useTranslation();
  const { formatDateTime } = useDateFormat();
  const { target, detail } = useAuditLogFormatter();

  if (logs.length === 0) {
    return (
      <p className="m-0 rounded-[10px] border border-border px-3 py-6 text-center text-caption text-muted">
        {t("admin.audit.empty")}
      </p>
    );
  }
  return (
    <div className={tableClassNames.wrapper}>
      <table className={tableClassNames.table}>
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column} scope="col" className={tableClassNames.header}>
                {t(`admin.audit.columns.${column}`)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => {
            const createdAt = toDate(log.createdAt);
            return (
              <tr key={log.id} className={tableClassNames.row}>
                <td className={cn(tableClassNames.cell, tableClassNames.numeric)}>
                  <time dateTime={createdAt.toISOString()}>{formatDateTime(createdAt)}</time>
                </td>
                <td className={tableClassNames.cell}>
                  {log.actor?.displayName ?? (
                    <span className="text-subtle">{t("admin.audit.unknownActor")}</span>
                  )}
                </td>
                <td className={tableClassNames.cell}>
                  <span
                    className={cn(
                      "rounded-sm border border-border bg-sunken px-1.5 py-px text-[11.5px] font-semibold text-muted",
                      sensitiveAuditActions.has(log.action) &&
                        "border-[color-mix(in_srgb,var(--c-danger)_40%,transparent)] text-danger",
                    )}
                  >
                    {t(`admin.audit.actions.${auditActionKeys[log.action]}`)}
                  </span>
                </td>
                <td className={cn(tableClassNames.cell, "max-w-56 truncate")}>{target(log)}</td>
                <td className={cn(tableClassNames.cell, "text-muted")}>{detail(log)}</td>
                <td className={cn(tableClassNames.cell, tableClassNames.numeric, "text-muted")}>
                  {log.ipAddress || "—"} ·{" "}
                  {summarizeUserAgent(log.userAgent) ?? t("admin.members.unknownDevice")}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};
