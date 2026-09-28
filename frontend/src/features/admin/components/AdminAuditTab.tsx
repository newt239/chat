import { useState } from "react";

import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { formatNumber } from "@chat/i18n";
import { IconChevronLeft, IconChevronRight, IconDownload } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { IconButton } from "#/components/ui/IconButton";
import { Select } from "#/components/ui/Select";
import { cn } from "#/components/ui/styles";
import { toast } from "#/components/ui/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { useAuditLogs } from "#/features/admin/hooks/useAdminQueries";
import { downloadText } from "#/features/admin/utils/downloadText";
import { auditActionKeys } from "#/features/admin/utils/labels";
import { AuditAction } from "#/gen/chat/v1/admin_service_pb";
import { preferencesAtom } from "#/providers/store/preferences";

import { AuditLogTable } from "./AuditLogTable";

import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

const PAGE_SIZE = 50;
const HOUR = 60 * 60 * 1000;

const periods = {
  all: null,
  day: 24 * HOUR,
  month: 30 * 24 * HOUR,
  quarter: 90 * 24 * HOUR,
  week: 7 * 24 * HOUR,
} as const;
type Period = keyof typeof periods;
const periodValues = ["day", "week", "month", "quarter", "all"] as const;

const actions = [
  AuditAction.LOGIN,
  AuditAction.LOGIN_FAILED,
  AuditAction.MEMBER_ROLE_CHANGED,
  AuditAction.MEMBER_SUSPENDED,
  AuditAction.MEMBER_RESUMED,
  AuditAction.CHANNEL_CREATED,
  AuditAction.CHANNEL_DELETED,
  AuditAction.CHANNEL_ARCHIVED,
  AuditAction.CHANNEL_UNARCHIVED,
  AuditAction.PERMISSION_CHANGED,
  AuditAction.DATA_EXPORTED,
];

type AdminAuditTabProps = {
  workspaceId: string;
  members: readonly AdminMember[];
};

export const AdminAuditTab = ({ workspaceId, members }: AdminAuditTabProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const { exportAuditLogs } = useAdminActions();
  const [actorId, setActorId] = useState("all");
  const [actionKey, setActionKey] = useState("all");
  const [period, setPeriod] = useState<Period>("month");
  const [page, setPage] = useState(0);
  // 期間の起点は絞り込みを変えたときに固定し、再描画のたびにクエリが変わらないようにする
  const [now, setNow] = useState(() => Date.now());

  const periodMs = periods[period];
  const action = actions.find((value) => auditActionKeys[value] === actionKey);
  const filter = {
    actions: action === undefined ? [] : [action],
    actorId: actorId === "all" ? undefined : actorId,
    since: periodMs === null ? undefined : timestampFromDate(new Date(now - periodMs)),
    workspaceId,
  };
  const { data, error, isPlaceholderData } = useAuditLogs({
    ...filter,
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
  });

  const changeFilter =
    <T,>(setter: (value: T) => void) =>
    (value: T) => {
      setter(value);
      setPage(0);
      setNow(Date.now());
    };
  const total = data?.totalCount ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-wrap items-end gap-2">
        <Select
          label={t("admin.audit.actorFilter")}
          className="w-44"
          value={actorId}
          onChange={changeFilter(setActorId)}
          options={[
            { label: t("admin.audit.allActors"), value: "all" },
            ...members.map((member) => ({ label: member.displayName, value: member.userId })),
          ]}
        />
        <Select
          label={t("admin.audit.actionFilter")}
          className="w-52"
          value={actionKey}
          onChange={changeFilter(setActionKey)}
          options={[
            { label: t("admin.audit.allActions"), value: "all" },
            ...actions.map((value) => ({
              label: t(`admin.audit.actions.${auditActionKeys[value]}`),
              value: auditActionKeys[value],
            })),
          ]}
        />
        <Select
          label={t("admin.audit.periodFilter")}
          className="w-32"
          value={period}
          onChange={changeFilter(setPeriod)}
          options={periodValues.map((value) => ({
            label: t(`admin.audit.periods.${value}`),
            value,
          }))}
        />
        <Button
          variant="secondary"
          className="ml-auto"
          isPending={exportAuditLogs.isPending}
          onPress={() => {
            exportAuditLogs.mutate(filter, {
              onSuccess: ({ content, fileName }) => {
                downloadText(content, fileName, "text/csv;charset=utf-8");
                toast(t("admin.audit.exported"), { tone: "success" });
              },
            });
          }}
        >
          <IconDownload aria-hidden />
          {t("admin.audit.export")}
        </Button>
      </div>
      {error ? (
        <p role="alert" className="m-0 text-caption text-danger">
          {t("admin.loadFailed")}
        </p>
      ) : (
        <div
          className={cn(
            "transition-opacity motion-reduce:transition-none",
            isPlaceholderData && "opacity-60",
          )}
        >
          <AuditLogTable logs={data?.logs ?? []} />
        </div>
      )}
      <nav className="flex items-center justify-between gap-2 text-xs text-muted tabular-nums">
        <span>
          {total === 0
            ? t("admin.audit.range", { from: 0, to: 0, total: 0 })
            : t("admin.audit.range", {
                from: formatNumber(page * PAGE_SIZE + 1, locale),
                to: formatNumber(Math.min(total, (page + 1) * PAGE_SIZE), locale),
                total: formatNumber(total, locale),
              })}
        </span>
        <span className="flex items-center gap-1">
          <IconButton
            label={t("admin.audit.prev")}
            isDisabled={page === 0}
            onPress={() => {
              setPage(page - 1);
            }}
          >
            <IconChevronLeft />
          </IconButton>
          <IconButton
            label={t("admin.audit.next")}
            isDisabled={page + 1 >= pageCount}
            onPress={() => {
              setPage(page + 1);
            }}
          >
            <IconChevronRight />
          </IconButton>
        </span>
      </nav>
    </div>
  );
};
