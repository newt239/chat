// 管理画面の表で共通の見た目
export const tableClassNames = {
  cell: "border-b border-border px-3 py-2 align-middle whitespace-nowrap",
  header:
    "sticky top-0 z-[1] border-b border-border bg-sunken px-3 py-2 text-left text-[11.5px] font-semibold whitespace-nowrap text-muted",
  numeric: "font-mono text-xs tabular-nums",
  row: "[&:last-child>td]:border-b-0",
  table: "w-full border-collapse font-sans text-[13px]",
  wrapper: "overflow-auto rounded-[10px] border border-border bg-surface",
};
