import { useState } from "react";
import type { ReactNode } from "react";

import { ToggleButton } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";

import { DataTable } from "./DataTable";

import type { TableData } from "./DataTable";

type ChartCardProps = {
  title: string;
  note: string;
  table: TableData;
  children: ReactNode;
  // 2 列のグリッドで横幅いっぱいに広げる
  wide?: boolean;
};

// どのグラフも表に切り替えられるようにし、ホバーできない環境でも値を読めるようにする
export const ChartCard = ({ title, note, table, children, wide }: ChartCardProps) => {
  const { t } = useTranslation();
  const [showTable, setShowTable] = useState(false);
  return (
    <section
      className={cn(
        "flex min-w-0 flex-col gap-3 rounded-xl border border-border bg-surface px-4 py-3.5 max-md:p-3",
        wide && "col-span-full",
      )}
    >
      <header className="flex flex-wrap items-baseline gap-x-2.5 gap-y-1">
        <h2 className="m-0 text-[13.5px] font-bold">{title}</h2>
        <span className="flex-1 text-[11.5px] text-muted">{note}</span>
        <ToggleButton
          isSelected={showTable}
          onChange={setShowTable}
          className={`cursor-pointer rounded-[6px] border border-border bg-transparent px-2 py-0.5 font-sans text-[11.5px] font-semibold text-accent-text data-hovered:bg-hover ${focusRing}`}
        >
          {showTable ? t("insights.table.showChart") : t("insights.table.showTable")}
        </ToggleButton>
      </header>
      {showTable ? <DataTable columns={table.columns} rows={table.rows} /> : children}
    </section>
  );
};
