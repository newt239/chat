import { useEffect } from "react";

import { createFileRoute, redirect } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ResponsiveLayout } from "@/features/layout/components/ResponsiveLayout";
import { store } from "@/providers/store";
import { isAuthenticatedAtom } from "@/providers/store/auth";
import { setIsChannelPageAtom } from "@/providers/store/ui";

const AppComponent = () => {
  const setIsChannelPage = useSetAtom(setIsChannelPageAtom);

  useEffect(() => {
    // チャンネルページでないことを設定
    setIsChannelPage(false);
  }, [setIsChannelPage]);

  return <ResponsiveLayout />;
};

export const Route = createFileRoute("/app")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppComponent,
});
