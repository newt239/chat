import type { ReactNode } from "react";

import { BackButton } from "#/components/block/BackButton/BackButton";

type PageHeaderProps = {
  icon: ReactNode;
  title: string;
  // 右端に並べる操作
  children?: ReactNode;
};

// スレッド一覧やメンションなど、チャンネル以外の画面の見出し。モバイルで積んだ画面では「戻る」を出す
export const PageHeader = ({ icon, title, children }: PageHeaderProps) => (
  <header className="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4.5 max-md:px-3 [&>svg]:size-4 [&>svg]:text-muted">
    <BackButton />
    {icon}
    <h1 className="m-0 min-w-0 flex-1 truncate text-title font-bold">{title}</h1>
    {children}
  </header>
);
