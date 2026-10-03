// 管理画面の表で共通の見た目
export const tableClassNames = {
  cell: "border-b border-border px-3 py-2 align-middle whitespace-nowrap",
  header:
    "sticky top-0 z-1 border-b border-border bg-sunken px-3 py-2 text-left text-caption font-semibold whitespace-nowrap text-muted",
  numeric: "font-mono text-xs tabular-nums",
  row: "[&:last-child>td]:border-b-0",
  table: "w-full border-collapse font-sans text-body-sm",
  wrapper: "overflow-auto rounded-lg border border-border bg-surface",
};
