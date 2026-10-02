import { formatBytes, formatDate, formatMonthDay, formatNumber, formatWeekday } from "@chat/i18n";
import { IconHash, IconLock, IconShieldCheck } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { useInsights } from "#/features/insights/hooks/useInsights";
import { isoWeekdayLabel, parseLocalDate, toHeatmapGrid } from "#/features/insights/utils/chart";
import { StorageCategory } from "#/gen/chat/v1/insight_service_pb";
import { useMyWorkspaceRole } from "#/hooks/useMyWorkspaceRole";
import { usePreferences } from "#/hooks/usePreferences";
import { isAdminRole } from "#/lib/isAdminRole";

import { BarChart } from "./BarChart";
import { ChartCard } from "./ChartCard";
import { HBarList } from "./HBarList";
import { Heatmap } from "./Heatmap";
import { InsightsKpis } from "./InsightsKpis";

import type { ChannelActivity } from "#/gen/chat/v1/insight_service_pb";

const storageCategoryKeys = {
  [StorageCategory.UNSPECIFIED]: "unspecified",
  [StorageCategory.IMAGE]: "image",
  [StorageCategory.VIDEO]: "video",
  [StorageCategory.AUDIO]: "audio",
  [StorageCategory.FILE]: "file",
} as const;

const TOP_CHANNELS = 8;
const POPULAR_CHANNELS = 5;

type InsightsDashboardProps = {
  workspaceId: string;
};

export const InsightsDashboard = ({ workspaceId }: InsightsDashboardProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  // 集計の日付はサーバーがプロフィールのタイムゾーンで区切った暦日で、端末の 0 時として読むため端末のタイムゾーンで書式化する
  const { data: insights, error } = useInsights(workspaceId);
  const { data: myRole } = useMyWorkspaceRole(workspaceId);

  if (error) {
    return (
      <p role="alert" className="m-0 p-6 text-caption text-danger">
        {t("insights.loadFailed")}
      </p>
    );
  }
  if (insights === undefined) {
    return (
      <div className="flex flex-col gap-3.5 p-[18px]">
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-56 w-full rounded-xl" />
        <Skeleton className="h-56 w-full rounded-xl" />
      </div>
    );
  }

  const count = (value: number) => formatNumber(value, locale);
  const countTooltip = (label: string, value: number, isPartial: boolean) =>
    `${t("insights.tooltip.count", { label, value: count(value) })}${isPartial ? t("insights.tooltip.partial") : ""}`;
  const toBars = (days: readonly { date: string; value: number }[]) =>
    days.map((day, index) => {
      const date = parseLocalDate(day.date);
      const isPartial = index === days.length - 1;
      return {
        isPartial,
        key: day.date,
        label: formatMonthDay(date, locale, undefined),
        tooltip: countTooltip(
          `${formatDate(date, locale, undefined)} (${formatWeekday(date, locale, undefined)})`,
          day.value,
          isPartial,
        ),
        value: day.value,
      };
    });
  const dateLabel = (value: string) => {
    const date = parseLocalDate(value);
    return `${formatDate(date, locale, undefined)} (${formatWeekday(date, locale, undefined)})`;
  };
  const channelRows = (
    channels: readonly ChannelActivity[],
    valueOf: (c: ChannelActivity) => number,
  ) =>
    channels.map((channel) => ({
      icon: channel.isPrivate ? <IconLock aria-hidden /> : <IconHash aria-hidden />,
      key: channel.channelId,
      label: channel.name,
      value: valueOf(channel),
      valueLabel: count(valueOf(channel)),
    }));

  const heatmapGrid = toHeatmapGrid(insights.heatmap).map((row) =>
    row.map((value) => value / Math.max(1, insights.heatmapWeeks)),
  );
  const popular = insights.channels
    .toSorted((a, b) => b.recentMessageCount - a.recentMessageCount)
    .filter((channel) => channel.recentMessageCount > 0)
    .slice(0, POPULAR_CHANNELS);
  const storage = insights.storageBreakdown.toSorted((a, b) => Number(b.bytes - a.bytes));
  const storageTotal = storage.reduce((sum, usage) => sum + Number(usage.bytes), 0);
  const categoryLabel = (category: StorageCategory) =>
    t(`insights.charts.storage.categories.${storageCategoryKeys[category]}`);

  return (
    <div className="flex flex-col gap-3.5 px-[18px] pt-4 pb-6 max-md:px-3.5 max-md:pt-3">
      <InsightsKpis insights={insights} />
      <div className="grid grid-cols-1 gap-3.5 lg:grid-cols-2">
        <ChartCard
          wide
          title={t("insights.charts.daily.title")}
          note={t("insights.charts.daily.note")}
          table={{
            columns: [
              t("insights.table.date"),
              t("insights.table.count"),
              t("insights.table.active"),
            ],
            rows: insights.dailyActivity.map((day) => [
              dateLabel(day.date),
              count(day.messageCount),
              count(day.activeMemberCount),
            ]),
          }}
        >
          <BarChart
            ariaLabel={t("insights.charts.daily.title")}
            height={160}
            labelEvery={7}
            data={toBars(
              insights.dailyActivity.map((day) => ({ date: day.date, value: day.messageCount })),
            )}
          />
        </ChartCard>
        <ChartCard
          title={t("insights.charts.channels.title")}
          note={t("insights.charts.channels.note")}
          table={{
            columns: [t("insights.table.channel"), t("insights.table.count")],
            rows: insights.channels.map((channel) => [channel.name, count(channel.messageCount)]),
          }}
        >
          <HBarList
            emptyLabel={t("insights.empty")}
            rows={channelRows(
              insights.channels
                .filter((channel) => channel.messageCount > 0)
                .slice(0, TOP_CHANNELS),
              (channel) => channel.messageCount,
            )}
          />
        </ChartCard>
        <ChartCard
          title={t("insights.charts.popular.title")}
          note={t("insights.charts.popular.note")}
          table={{
            columns: [t("insights.table.channel"), t("insights.table.count")],
            rows: popular.map((channel) => [channel.name, count(channel.recentMessageCount)]),
          }}
        >
          <HBarList
            emptyLabel={t("insights.empty")}
            rows={channelRows(popular, (channel) => channel.recentMessageCount)}
          />
        </ChartCard>
        <ChartCard
          wide
          title={t("insights.charts.heatmap.title")}
          note={t("insights.charts.heatmap.note", { weeks: insights.heatmapWeeks })}
          table={{
            columns: [
              t("insights.table.weekday"),
              ...Array.from({ length: 24 }, (_, hour) => String(hour)),
            ],
            rows: heatmapGrid.map((row, index) => [
              isoWeekdayLabel(index + 1, locale),
              ...row.map((value) => count(value)),
            ]),
          }}
        >
          <Heatmap grid={heatmapGrid} ariaLabel={t("insights.charts.heatmap.title")} />
        </ChartCard>
        <ChartCard
          title={t("insights.charts.storage.title")}
          note={t("insights.charts.storage.note", { total: formatBytes(storageTotal, locale) })}
          table={{
            columns: [
              t("insights.table.category"),
              t("insights.table.size"),
              t("insights.table.files"),
            ],
            rows: storage.map((usage) => [
              categoryLabel(usage.category),
              formatBytes(Number(usage.bytes), locale),
              count(usage.fileCount),
            ]),
          }}
        >
          <HBarList
            emptyLabel={t("insights.empty")}
            rows={storage.map((usage) => ({
              icon: null,
              key: String(usage.category),
              label: categoryLabel(usage.category),
              value: Number(usage.bytes),
              valueLabel: formatBytes(Number(usage.bytes), locale),
            }))}
          />
        </ChartCard>
        <ChartCard
          title={t("insights.charts.mine.title")}
          note={t("insights.charts.mine.note")}
          table={{
            columns: [t("insights.table.date"), t("insights.table.count")],
            rows: insights.myDailyMessages.map((day) => [dateLabel(day.date), count(day.count)]),
          }}
        >
          <BarChart
            ariaLabel={t("insights.charts.mine.title")}
            height={120}
            labelEvery={1}
            data={toBars(
              insights.myDailyMessages.map((day) => ({ date: day.date, value: day.count })),
            )}
          />
        </ChartCard>
      </div>
      <p className="m-0 flex flex-wrap items-center gap-1.5 text-xs text-muted">
        <IconShieldCheck aria-hidden className="size-3.5" />
        {t("insights.note")}
        {isAdminRole(myRole) && (
          <Link to="/app/$workspaceId/admin" params={{ workspaceId }}>
            {t("insights.openAdmin")}
          </Link>
        )}
      </p>
    </div>
  );
};
