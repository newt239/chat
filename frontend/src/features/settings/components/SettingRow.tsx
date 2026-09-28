import type { ReactNode } from "react";

type SettingRowProps = {
  title: string;
  description: string | null;
  children: ReactNode;
};

// 設定の 1 行。左に項目名と説明、右に操作を置く（狭い画面では折り返す）
export const SettingRow = ({ title, description, children }: SettingRowProps) => (
  <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-border py-3 [word-break:auto-phrase] last:border-b-0">
    <div className="flex min-w-36 flex-1 flex-col">
      <b className="text-body-strong">{title}</b>
      {description && <span className="text-caption text-muted">{description}</span>}
    </div>
    {children}
  </div>
);
