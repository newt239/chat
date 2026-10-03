import { cn, focusRing } from "#/components/ui/styles/styles";

// 選択中は data-selected（ToggleButton）か data-active、名前を解決できなければ data-invalid
export const chipClassName = cn(
  "inline-flex h-7.5 shrink-0 cursor-pointer items-center gap-1 rounded-full border border-border-strong bg-surface px-3 font-sans text-body-sm whitespace-nowrap text-muted data-hovered:bg-hover [&_svg]:size-3.5",
  "data-selected:border-accent data-selected:bg-accent-soft data-selected:font-semibold data-selected:text-accent-text",
  "data-[active=true]:border-accent data-[active=true]:bg-accent-soft data-[active=true]:font-semibold data-[active=true]:text-accent-text",
  "data-[invalid=true]:border-danger data-[invalid=true]:text-danger",
  focusRing,
);
