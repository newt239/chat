import type { ReactNode } from "react";

import { cn } from "#/components/ui/styles/styles";

const tones = {
  // 強調したいラベル
  accent: "rounded-[6px] bg-accent-soft px-1.5 text-xs leading-5 font-semibold text-accent-text",
  // 未読数などの件数
  count:
    "h-[17px] min-w-[18px] justify-center rounded-full bg-badge px-[5px] text-[10.5px] font-bold text-badge-fg tabular-nums",
  // BOT・管理者などのラベル
  tag: "rounded-sm border border-border bg-sunken px-1 text-[9.5px] leading-[15px] font-bold tracking-[.06em] text-muted",
};

type BadgeProps = {
  children: ReactNode;
  tone?: keyof typeof tones;
  className?: string;
};

export const Badge = ({ children, tone = "count", className }: BadgeProps) => (
  <span className={cn("inline-flex shrink-0 items-center font-sans", tones[tone], className)}>
    {children}
  </span>
);
