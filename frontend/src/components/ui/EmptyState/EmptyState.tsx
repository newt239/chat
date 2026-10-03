import type { ReactNode } from "react";

type EmptyStateProps = {
  icon: ReactNode;
  title: string;
  description: string;
};

// 一覧が空のときや、まだ中身のない画面に出す案内
export const EmptyState = ({ icon, title, description }: EmptyStateProps) => (
  <div className="m-auto flex flex-col items-center gap-1.5 p-6 text-center font-sans text-muted [word-break:auto-phrase]">
    <span className="mb-1 grid size-11 place-items-center rounded-lg bg-accent-soft text-accent-text [&_svg]:size-5.5">
      {icon}
    </span>
    <b className="text-title text-text">{title}</b>
    <span className="max-w-80 text-body-sm">{description}</span>
  </div>
);
