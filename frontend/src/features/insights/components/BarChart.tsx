import { useState } from "react";

import { formatNumber } from "@chat/i18n";
import { useAtomValue } from "jotai";
import { motion } from "motion/react";

import { cn } from "#/components/ui/styles";
import { niceMax } from "#/features/insights/utils/chart";
import { transitions } from "#/lib/motion";
import { preferencesAtom } from "#/providers/store/preferences";

import { ChartTooltip } from "./ChartTooltip";

export type BarDatum = {
  key: string;
  // 軸に出す短い表示
  label: string;
  // ホバー時に出す説明
  tooltip: string;
  value: number;
  // 今日のように集計途中の値は薄く描く
  isPartial: boolean;
};

type BarChartProps = {
  data: readonly BarDatum[];
  height: number;
  // 軸のラベルを何本おきに出すか
  labelEvery: number;
  ariaLabel: string;
};

export const BarChart = ({ data, height, labelEvery, ariaLabel }: BarChartProps) => {
  const { locale } = useAtomValue(preferencesAtom);
  const [hovered, setHovered] = useState<number | null>(null);
  const max = niceMax(Math.max(0, ...data.map((datum) => datum.value)));
  const percentOf = (value: number) => (value / max) * 100;
  const hoveredDatum = hovered === null ? undefined : data[hovered];
  const xOf = (index: number) => ((index + 0.5) / data.length) * 100;

  return (
    <div
      role="img"
      aria-label={ariaLabel}
      className="grid grid-cols-[auto_1fr] gap-x-2 text-subtle"
    >
      <div
        aria-hidden
        className="-my-[5px] flex flex-col justify-between text-right font-mono text-[10.5px] leading-none tabular-nums"
        style={{ height }}
      >
        <span>{formatNumber(max, locale)}</span>
        <span>{formatNumber(max / 2, locale)}</span>
        <span>0</span>
      </div>
      <div
        className="relative border-b border-border-strong"
        style={{ height }}
        onPointerLeave={() => {
          setHovered(null);
        }}
      >
        <span aria-hidden className="absolute inset-x-0 top-0 border-t border-border" />
        <span aria-hidden className="absolute inset-x-0 top-1/2 border-t border-border" />
        <div className="absolute inset-0 flex items-end gap-0.5">
          {data.map((datum, index) => (
            <div
              key={datum.key}
              className="flex h-full min-w-0 flex-1 items-end justify-center"
              onPointerEnter={() => {
                setHovered(index);
              }}
            >
              <motion.i
                className={cn(
                  "block w-full max-w-6 rounded-t-sm bg-accent",
                  datum.value > 0 && "min-h-0.5",
                  hovered === index && "bg-accent-hover",
                  datum.isPartial && "opacity-45",
                )}
                initial={{ height: 0 }}
                animate={{ height: `${percentOf(datum.value)}%` }}
                transition={transitions.spring}
              />
            </div>
          ))}
        </div>
        {hoveredDatum !== undefined && hovered !== null && (
          <ChartTooltip
            x={xOf(hovered)}
            y={100 - percentOf(hoveredDatum.value)}
            text={hoveredDatum.tooltip}
          />
        )}
      </div>
      <span />
      <div aria-hidden className="relative h-[18px] font-mono text-[10.5px]">
        {data.map((datum, index) =>
          (index % labelEvery === 0 && index < data.length - Math.ceil(labelEvery / 2)) ||
          index === data.length - 1 ? (
            <span
              key={datum.key}
              className={cn(
                "absolute top-1 -translate-x-1/2 whitespace-nowrap",
                index === data.length - 1 && "-translate-x-[80%]",
              )}
              style={{ left: `${xOf(index)}%` }}
            >
              {datum.label}
            </span>
          ) : null,
        )}
      </div>
    </div>
  );
};
