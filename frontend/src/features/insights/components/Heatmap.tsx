import { useState } from "react";

import { formatNumber } from "@chat/i18n/format";
import { useTranslation } from "react-i18next";

import { cn } from "#/components/ui/styles/styles";
import { isoWeekdayLabel } from "#/features/insights/utils/chart";
import { usePreferences } from "#/hooks/usePreferences";

type HeatmapProps = {
  // 7（月〜日）× 24 時間の平均値
  grid: readonly (readonly number[])[];
  ariaLabel: string;
};

const HOURS = Array.from({ length: 24 }, (_, hour) => hour);
const HOUR_TICKS = [0, 6, 12, 18, 23];
const LEGEND_STEPS = [0, 0.25, 0.5, 0.75, 1];

// アクセントの 1 色を面の色と混ぜて濃淡を作る。ダークでは面が暗いので多いほど明るくなる
const colorOf = (ratio: number) =>
  ratio <= 0
    ? "var(--c-sunken)"
    : `color-mix(in oklab, var(--c-accent) ${Math.round(12 + 88 * ratio)}%, var(--c-sunken))`;

export const Heatmap = ({ grid, ariaLabel }: HeatmapProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const [hovered, setHovered] = useState<{ weekday: number; hour: number } | null>(null);
  const max = Math.max(0, ...grid.flat());
  const hoveredValue = hovered === null ? undefined : grid[hovered.weekday - 1]?.[hovered.hour];

  return (
    <div className="flex flex-col gap-2">
      <div
        role="img"
        aria-label={ariaLabel}
        className="grid grid-cols-[20px_repeat(24,minmax(0,1fr))] gap-0.5 font-mono text-[10px] text-subtle"
        onPointerLeave={() => {
          setHovered(null);
        }}
      >
        {grid.map((row, rowIndex) => {
          const weekday = rowIndex + 1;
          return [
            <span key={`label-${weekday}`} className="flex items-center">
              {isoWeekdayLabel(weekday, locale)}
            </span>,
            ...HOURS.map((hour) => (
              <i
                key={`${weekday}-${hour}`}
                className={cn(
                  "block aspect-square min-w-0 rounded-[2px]",
                  hovered?.weekday === weekday &&
                    hovered.hour === hour &&
                    "outline-2 -outline-offset-1 outline-text outline-solid",
                )}
                style={{ background: colorOf(max === 0 ? 0 : (row[hour] ?? 0) / max) }}
                onPointerEnter={() => {
                  setHovered({ hour, weekday });
                }}
              />
            )),
          ];
        })}
        <span />
        <div className="col-start-2 -col-end-1 flex justify-between pt-0.5">
          {HOUR_TICKS.map((hour) => (
            <span key={hour}>{hour}</span>
          ))}
        </div>
      </div>
      <div className="flex flex-wrap justify-between gap-2.5 text-[11.5px] text-muted tabular-nums">
        <span aria-live="polite">
          {hovered === null || hoveredValue === undefined
            ? t("insights.charts.heatmap.hint")
            : t("insights.charts.heatmap.cell", {
                hour: hovered.hour,
                value: formatNumber(hoveredValue, locale),
                weekday: isoWeekdayLabel(hovered.weekday, locale),
              })}
        </span>
        <span aria-hidden className="flex items-center gap-[3px] text-[11px]">
          {t("insights.charts.heatmap.less")}
          {LEGEND_STEPS.map((step) => (
            <i
              key={step}
              className="block h-2.5 w-3.5 rounded-[2px]"
              style={{ background: colorOf(step) }}
            />
          ))}
          {t("insights.charts.heatmap.more")}
        </span>
      </div>
    </div>
  );
};
