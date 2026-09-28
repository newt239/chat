import { formatBytes, formatNumber } from "@chat/i18n";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { directionOf, percentChange, signed } from "#/features/insights/utils/chart";
import { preferencesAtom } from "#/providers/store/preferences";

import { KpiCard } from "./KpiCard";

import type { KpiDelta } from "./KpiCard";

import type { GetInsightsResponse } from "#/gen/chat/v1/insight_service_pb";

type InsightsKpisProps = {
  insights: GetInsightsResponse;
};

export const InsightsKpis = ({ insights }: InsightsKpisProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);

  const percentDelta = (
    current: bigint | undefined,
    previous: bigint | undefined,
    isGoodWhenUp: boolean | null,
  ): KpiDelta | null => {
    const percent = percentChange(Number(current ?? 0n), Number(previous ?? 0n));
    return percent === null
      ? null
      : {
          direction: directionOf(percent),
          isGoodWhenUp,
          text: t("insights.kpi.percentDelta", {
            value: signed(formatNumber(percent, locale), percent),
          }),
        };
  };

  const rate = insights.activeRate;
  const ratePoints = Math.round(((rate?.current ?? 0) - (rate?.previous ?? 0)) * 100);
  const storage = insights.storageBytes;
  const storageDiff = Number((storage?.current ?? 0n) - (storage?.previous ?? 0n));
  const myTotal = insights.myDailyMessages.reduce((sum, day) => sum + day.count, 0);

  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(170px,1fr))] gap-2.5 max-md:grid-cols-2 max-md:gap-2">
      <KpiCard
        label={t("insights.kpi.activeMembers")}
        value={formatNumber(Number(insights.activeMembers?.current ?? 0n), locale)}
        unit={t("insights.kpi.ofMembers", {
          count: Number(insights.memberCount?.current ?? 0n),
        })}
        delta={percentDelta(
          insights.activeMembers?.current,
          insights.activeMembers?.previous,
          true,
        )}
      />
      <KpiCard
        label={t("insights.kpi.activeRate")}
        value={`${formatNumber(Math.round((rate?.current ?? 0) * 100), locale)}%`}
        unit=""
        delta={{
          direction: directionOf(ratePoints),
          isGoodWhenUp: true,
          text: t("insights.kpi.pointDelta", {
            value: signed(formatNumber(ratePoints, locale), ratePoints),
          }),
        }}
      />
      <KpiCard
        label={t("insights.kpi.messages")}
        value={formatNumber(Number(insights.messageCount?.current ?? 0n), locale)}
        unit={t("insights.kpi.unitCount")}
        delta={percentDelta(insights.messageCount?.current, insights.messageCount?.previous, null)}
      />
      <KpiCard
        label={t("insights.kpi.storage")}
        value={formatBytes(Number(storage?.current ?? 0n), locale)}
        unit=""
        delta={{
          direction: directionOf(storageDiff),
          isGoodWhenUp: null,
          text: t("insights.kpi.storageDelta", {
            value: `${storageDiff < 0 ? "-" : "+"}${formatBytes(Math.abs(storageDiff), locale)}`,
          }),
        }}
      />
      <KpiCard
        label={t("insights.kpi.myMessages")}
        value={formatNumber(myTotal, locale)}
        unit={t("insights.kpi.unitCount")}
        delta={null}
      />
    </div>
  );
};
