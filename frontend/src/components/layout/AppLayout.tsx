import type { ReactNode } from "react";

import { Header } from "@/components/layout/Header";
import { AuthGuard } from "@/features/auth/components/AuthGuard";

type AppLayoutProps = {
  children: ReactNode;
}

export const AppLayout = ({ children }: AppLayoutProps) => {
  return (
    <AuthGuard>
      <div className="h-full flex flex-col bg-gray-50">
        <Header />
        <div className="flex-1 min-h-0 p-6">
          <div className="h-full min-h-0">{children}</div>
        </div>
      </div>
    </AuthGuard>
  );
};
