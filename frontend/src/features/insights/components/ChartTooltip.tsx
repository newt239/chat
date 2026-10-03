import { cn } from "#/components/ui/styles/styles";

type ChartTooltipProps = {
  // プロット領域に対する位置（%）
  x: number;
  y: number;
  text: string;
};

// 端では左右にずらしてプロットからはみ出さないようにする
export const ChartTooltip = ({ x, y, text }: ChartTooltipProps) => (
  <div
    role="status"
    className={cn(
      "pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-[calc(100%+8px)] rounded-md bg-text px-2 py-1 font-sans text-caption whitespace-nowrap text-surface tabular-nums shadow-md",
      x > 82 && "-translate-x-[calc(100%-8px)]",
      x < 18 && "-translate-x-2",
    )}
    style={{ left: `${x}%`, top: `${y}%` }}
  >
    {text}
  </div>
);
