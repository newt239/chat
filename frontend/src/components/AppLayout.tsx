import type { ReactNode } from "react";

import { Header } from "@/components/Header";
import { AuthGuard } from "@/features/auth/components/AuthGuard";

interface AppLayoutProps {
  children: ReactNode;
}

export const AppLayout = ({ children }: AppLayoutProps) => {
  return (
    <AuthGuard>
      <div className="h-full flex flex-col bg-gray-50">
        <Header />
        <div className="flex-1 p-6">{children}</div>
      </div>
    </AuthGuard>
  );
};
