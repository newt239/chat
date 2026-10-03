import { useState } from "react";
import type { ReactNode } from "react";

import { motion } from "motion/react";

import { cn } from "#/components/ui/styles/styles";
import { transitions } from "#/lib/motion";

type HBarRow = {
  key: string;
  label: string;
  // ラベルの前に置くアイコン（チャンネルの種類など）
  icon: ReactNode;
  value: number;
  valueLabel: string;
};

type HBarListProps = {
  rows: readonly HBarRow[];
  emptyLabel: string;
};

// 値はバーの右に常に出し、ホバーでは行を強調する
export const HBarList = ({ rows, emptyLabel }: HBarListProps) => {
  const [hovered, setHovered] = useState<string | null>(null);
  const max = Math.max(0, ...rows.map((row) => row.value));
  if (rows.length === 0) {
    return <p className="m-0 py-4 text-center text-caption text-muted">{emptyLabel}</p>;
  }
  return (
    <ul
      className="m-0 flex list-none flex-col gap-1.75 p-0"
      onPointerLeave={() => {
        setHovered(null);
      }}
    >
      {rows.map((row) => (
        <li
          key={row.key}
          className="grid grid-cols-[minmax(80px,150px)_1fr_72px] items-center gap-2.5 text-label font-normal"
          onPointerEnter={() => {
            setHovered(row.key);
          }}
        >
          <span className="flex min-w-0 items-center gap-1 text-text [&_svg]:size-3.25 [&_svg]:shrink-0 [&_svg]:text-muted">
            {row.icon}
            <span className="truncate">{row.label}</span>
          </span>
          <span className="relative h-2.5">
            <motion.i
              className={cn(
                "absolute inset-y-0 left-0 block rounded-r-sm bg-accent",
                row.value > 0 && "min-w-0.5",
                hovered === row.key && "bg-accent-hover",
              )}
              initial={{ width: 0 }}
              animate={{ width: `${max === 0 ? 0 : (row.value / max) * 100}%` }}
              transition={transitions.spring}
            />
          </span>
          <span className="text-right font-mono text-xs text-muted tabular-nums">
            {row.valueLabel}
          </span>
        </li>
      ))}
    </ul>
  );
};
