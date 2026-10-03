import type { ReactNode } from "react";

import { create } from "@bufbuild/protobuf";
import { Code, ConnectError, createRouterTransport } from "@connectrpc/connect";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { createStore, Provider as JotaiProvider } from "jotai";
import { describe, expect, test, vi } from "vite-plus/test";

import { ColorMode, UserPreferencesSchema } from "#/gen/chat/v1/user_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { sessionAtom } from "#/providers/store/auth";
import { storedPreferencesAtom } from "#/providers/store/preferences";

import { usePreferences, useUpdatePreferences } from "./usePreferences";

import type { ConnectRouter } from "@connectrpc/connect";

vi.mock("#/components/ui/ToastRegion/toast", () => ({ toast: vi.fn() }));

const setup = (hasSession: boolean, routes: (router: ConnectRouter) => void) => {
  const store = createStore();
  store.set(storedPreferencesAtom, create(UserPreferencesSchema, { colorMode: ColorMode.LIGHT }));
  store.set(sessionAtom, hasSession ? { accessToken: "a", userId: "u1" } : null);
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <JotaiProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <TransportProvider transport={createRouterTransport(routes)}>{children}</TransportProvider>
      </QueryClientProvider>
    </JotaiProvider>
  );
  return renderHook(() => ({ preferences: usePreferences(), ...useUpdatePreferences() }), {
    wrapper,
  });
};

const darkAccount = () => ({
  user: {
    id: "u1",
    preferences: create(UserPreferencesSchema, { colorMode: ColorMode.DARK, locale: "ja" }),
  },
});

describe("usePreferences", () => {
  test("ログイン前は端末に残した設定を使う", () => {
    const { result } = setup(false, () => {});
    expect(result.current.preferences.mode).toBe("light");
  });

  test("ログイン中はアカウントの設定を使う", async () => {
    const { result } = setup(true, (router) => {
      router.rpc(UserService.method.getMe, darkAccount);
    });
    await waitFor(() => {
      expect(result.current.preferences.mode).toBe("dark");
    });
  });

  test("保存に失敗したら元の設定に戻す", async () => {
    const { result } = setup(true, (router) => {
      router.rpc(UserService.method.getMe, darkAccount);
      router.rpc(UserService.method.updatePreferences, () => {
        throw new ConnectError("failed", Code.Internal);
      });
    });
    await waitFor(() => {
      expect(result.current.preferences.mode).toBe("dark");
    });

    act(() => {
      result.current.update({ mode: "light" });
    });
    expect(result.current.preferences.mode).toBe("light");
    await waitFor(() => {
      expect(result.current.preferences.mode).toBe("dark");
    });
  });
});
