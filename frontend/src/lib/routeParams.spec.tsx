import type { ReactNode } from "react";

import { renderHook } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { describe, expect, test } from "vite-plus/test";

import { useChannelId, useOptionalRouteParams, useWorkspaceId } from "#/lib/routeParams";

/** 指定した URL を `/app/:workspaceId/:channelId` パターンで解決するラッパーを作る。 */
const createWrapper = (initialPath: string, routePath: string) => {
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path={routePath} element={children} />
      </Routes>
    </MemoryRouter>
  );
  return Wrapper;
};

describe("useWorkspaceId", () => {
  test("ルートパラメータから workspaceId を取り出す", () => {
    const { result } = renderHook(() => useWorkspaceId(), {
      wrapper: createWrapper("/app/ws1", "/app/:workspaceId"),
    });
    expect(result.current).toBe("ws1");
  });

  test("workspaceId が無いルートでは throw する", () => {
    expect(() =>
      renderHook(() => useWorkspaceId(), {
        wrapper: createWrapper("/app", "/app"),
      }),
    ).toThrow("workspaceId がルートパラメータに存在しません");
  });
});

describe("useChannelId", () => {
  test("ルートパラメータから channelId を取り出す", () => {
    const { result } = renderHook(() => useChannelId(), {
      wrapper: createWrapper("/app/ws1/ch1", "/app/:workspaceId/:channelId"),
    });
    expect(result.current).toBe("ch1");
  });

  test("channelId が無いルートでは throw する", () => {
    expect(() =>
      renderHook(() => useChannelId(), {
        wrapper: createWrapper("/app/ws1", "/app/:workspaceId"),
      }),
    ).toThrow("channelId がルートパラメータに存在しません");
  });
});

describe("useOptionalRouteParams", () => {
  test("パラメータが無い階層でも throw せず undefined を返す", () => {
    const { result } = renderHook(() => useOptionalRouteParams(), {
      wrapper: createWrapper("/app", "/app"),
    });
    expect(result.current.workspaceId).toBeUndefined();
    expect(result.current.channelId).toBeUndefined();
  });

  test("存在するパラメータは取り出せる", () => {
    const { result } = renderHook(() => useOptionalRouteParams(), {
      wrapper: createWrapper("/app/ws1/ch1", "/app/:workspaceId/:channelId"),
    });
    expect(result.current.workspaceId).toBe("ws1");
    expect(result.current.channelId).toBe("ch1");
  });
});
