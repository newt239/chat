import { IconArrowDownRight, IconArrowUpRight, IconMinus } from "@tabler/icons-react";

import { cn } from "#/components/ui/styles";

import type { Direction } from "#/features/insights/utils/chart";

export type KpiDelta = {
  text: string;
  direction: Direction;
  // 増えるのが良いか（null は良し悪しを付けない）
  isGoodWhenUp: boolean | null;
};

type KpiCardProps = {
  label: string;
  value: string;
  unit: string;
  delta: KpiDelta | null;
};

const icons = { down: IconArrowDownRight, flat: IconMinus, up: IconArrowUpRight };

const toneOf = ({ direction, isGoodWhenUp }: KpiDelta) => {
  if (direction === "flat" || isGoodWhenUp === null) {
    return "text-muted";
  }
  return (direction === "up") === isGoodWhenUp ? "text-success" : "text-danger";
};

export const KpiCard = ({ label, value, unit, delta }: KpiCardProps) => {
  const Icon = delta === null ? null : icons[delta.direction];
  return (
    <div className="flex min-w-0 flex-col gap-0.5 rounded-xl border border-border bg-surface px-3.5 py-3 max-md:px-3 max-md:py-2.5">
      <span className="text-xs text-muted">{label}</span>
      <b className="text-2xl leading-tight font-bold tracking-[-0.01em] max-md:text-[19px]">
        {value}
        {unit !== "" && (
          <small className="ml-[3px] text-[12.5px] font-medium tracking-normal text-muted max-md:ml-0 max-md:block max-md:text-[11.5px]">
            {unit}
          </small>
        )}
      </b>
      {delta !== null && Icon !== null && (
        <span
          className={cn(
            "inline-flex items-center gap-[3px] text-[11.5px] font-semibold tabular-nums",
            toneOf(delta),
          )}
        >
          <Icon aria-hidden className="size-3.5" />
          {delta.text}
        </span>
      )}
    </div>
  );
};
