import type { ReactNode } from "react";

type AuthCardProps = {
  title: string;
  children: ReactNode;
  footer: ReactNode;
};

// ログインと新規登録で共有する枠
export const AuthCard = ({ title, children, footer }: AuthCardProps) => (
  <main className="flex min-h-full items-center justify-center bg-bg px-4 py-10 font-sans text-text">
    <section className="flex w-full max-w-sm flex-col gap-5 rounded-xl border border-border bg-surface p-6 shadow-md">
      <h1 className="m-0 text-center text-title">{title}</h1>
      {children}
      <p className="m-0 text-center text-caption text-muted">{footer}</p>
    </section>
  </main>
);
