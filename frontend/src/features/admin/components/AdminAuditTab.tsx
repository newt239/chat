import { useState } from "react";

import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { formatNumber } from "@chat/i18n/format";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { IconDownload } from "@tabler/icons-react";
import { keepPreviousData } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Select } from "#/components/ui/Select/Select";
import { cn } from "#/components/ui/styles/styles";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useAdminActions } from "#/features/admin/hooks/useAdminActions";
import { auditPeriodValues } from "#/features/admin/schemas";
import { downloadText } from "#/features/admin/utils/downloadText";
import { auditActions } from "#/features/admin/utils/labels";
import { AdminService } from "#/gen/chat/v1/admin_service_pb";
import { usePreferences } from "#/hooks/usePreferences";

import { AuditLogTable } from "./AuditLogTable";

import type { adminSearchSchema } from "#/features/admin/schemas";
import type { AdminMember } from "#/gen/chat/v1/admin_service_pb";

import type { z } from "zod";

type AdminSearch = z.infer<typeof adminSearchSchema>;

const PAGE_SIZE = 50;
const HOUR = 60 * 60 * 1000;

const presetDurations = {
  all: null,
  day: 24 * HOUR,
  month: 30 * 24 * HOUR,
  quarter: 90 * 24 * HOUR,
  week: 7 * 24 * HOUR,
} as const;

const adminRoute = getRouteApi("/app/$workspaceId/admin");

const pad = (value: number) => String(value).padStart(2, "0");

// <input type="datetime-local"> の形式（ブラウザのタイムゾーン、分まで）にする
const toLocalDateTime = (date: Date) =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;

const parseLocalDateTime = (value: string | undefined) =>
  value === undefined ? undefined : new Date(value);

// 終了は含まない（until 未満）
const auditRange = (search: AdminSearch, now: number) => {
  const period = search.period ?? "month";
  if (period === "custom") {
    return { since: parseLocalDateTime(search.since), until: parseLocalDateTime(search.until) };
  }
  const duration = presetDurations[period];
  return { since: duration === null ? undefined : new Date(now - duration), until: undefined };
};

type AdminAuditTabProps = {
  workspaceId: string;
  members: readonly AdminMember[];
};

export const AdminAuditTab = ({ workspaceId, members }: AdminAuditTabProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const { exportAuditLogs } = useAdminActions();
  const search = adminRoute.useSearch();
  const navigate = adminRoute.useNavigate();
  // プリセットの期間の起点は絞り込みを変えたときに固定し、再描画のたびにクエリが変わらないようにする
  const [now, setNow] = useState(() => Date.now());

  const period = search.period ?? "month";
  const { since, until } = auditRange(search, now);
  const isInvalidRange = since !== undefined && until !== undefined && since >= until;
  const action = auditActions.find((entry) => entry.key === search.action);

  const filter = {
    actions: action === undefined ? [] : [action.action],
    actorId: search.actor,
    since: since === undefined ? undefined : timestampFromDate(since),
    until: until === undefined ? undefined : timestampFromDate(until),
    workspaceId,
  };
  const { data, error, fetchNextPage, hasNextPage, isFetchingNextPage, isPlaceholderData } =
    // 保存先を NoSQL に移せるよう総件数やオフセットは使わず、ページトークンで続きを読み込む。絞り込みを変えても前の結果を表示したままにする
    useInfiniteQuery(
      AdminService.method.listAuditLogs,
      { ...filter, limit: PAGE_SIZE, pageToken: "" },
      {
        getNextPageParam: (res) => res.nextPageToken || undefined,
        pageParamKey: "pageToken",
        placeholderData: keepPreviousData,
      },
    );
  const logs = data?.pages.flatMap((page) => page.logs) ?? [];

  const updateSearch = (patch: Partial<AdminSearch>) => {
    setNow(Date.now());
    void navigate({ replace: true, search: (prev) => ({ ...prev, ...patch }) });
  };

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="flex flex-wrap items-start gap-2">
          <Select
            label={t("admin.audit.actorFilter")}
            className="w-44"
            value={search.actor ?? "all"}
            onChange={(value) => {
              updateSearch({ actor: value === "all" ? undefined : value });
            }}
            options={[
              { label: t("admin.audit.allActors"), value: "all" },
              ...members.map((member) => ({ label: member.displayName, value: member.userId })),
            ]}
          />
          <Select
            label={t("admin.audit.actionFilter")}
            className="w-52"
            value={action?.key ?? "all"}
            onChange={(value) => {
              updateSearch({ action: value === "all" ? undefined : value });
            }}
            options={[
              { label: t("admin.audit.allActions"), value: "all" },
              ...auditActions.map(({ key }) => ({
                label: t(`admin.audit.actions.${key}`),
                value: key,
              })),
            ]}
          />
          <Select
            label={t("admin.audit.periodFilter")}
            className="w-36"
            value={period}
            onChange={(value) => {
              // 日時の指定に切り替えたときは、それまでの期間を初期値にする
              updateSearch(
                value === "custom"
                  ? {
                      period: value,
                      since: toLocalDateTime(since ?? new Date(now - presetDurations.month)),
                      until: undefined,
                    }
                  : { period: value, since: undefined, until: undefined },
              );
            }}
            options={auditPeriodValues.map((value) => ({
              label: t(`admin.audit.periods.${value}`),
              value,
            }))}
          />
          {period === "custom" && (
            <>
              <TextField
                type="datetime-local"
                label={t("admin.audit.since")}
                className="w-52"
                value={search.since ?? ""}
                onChange={(value) => {
                  updateSearch({ since: value || undefined });
                }}
              />
              <TextField
                type="datetime-local"
                label={t("admin.audit.until")}
                className="w-52"
                value={search.until ?? ""}
                errorMessage={isInvalidRange ? t("admin.audit.invalidRange") : undefined}
                onChange={(value) => {
                  updateSearch({ until: value || undefined });
                }}
              />
            </>
          )}
        </div>
        <Button
          variant="secondary"
          className="mt-5.5 ml-auto"
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
          <AuditLogTable logs={logs} />
        </div>
      )}
      <nav className="flex items-center justify-between gap-2 text-xs text-muted tabular-nums">
        <span>{t("admin.audit.loaded", { shown: formatNumber(logs.length, locale) })}</span>
        {hasNextPage && (
          <Button
            variant="secondary"
            size="sm"
            isPending={isFetchingNextPage}
            onPress={() => {
              void fetchNextPage();
            }}
          >
            {t("admin.audit.loadMore")}
          </Button>
        )}
      </nav>
    </div>
  );
};
