import type { ReactNode } from "react";

import { act, renderHook } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, test } from "vite-plus/test";

import { useSearchQueryParams } from "#/features/search/hooks/useSearchQueryParams";

const createWrapper = (initialPath: string) => {
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  );
  return Wrapper;
};

describe("useSearchQueryParams", () => {
  test("クエリが無いときは既定値を返す", () => {
    const { result } = renderHook(() => useSearchQueryParams(), {
      wrapper: createWrapper("/app/ws1/search"),
    });
    expect(result.current.query).toEqual({ q: "", filter: "all", page: 1 });
  });

  test("URL のクエリを検証して返す", () => {
    const { result } = renderHook(() => useSearchQueryParams(), {
      wrapper: createWrapper("/app/ws1/search?q=hello&filter=messages&page=3"),
    });
    expect(result.current.query).toEqual({ q: "hello", filter: "messages", page: 3 });
  });

  test("不正な filter は all にフォールバックする", () => {
    const { result } = renderHook(() => useSearchQueryParams(), {
      wrapper: createWrapper("/app/ws1/search?filter=unknown"),
    });
    expect(result.current.query.filter).toBe("all");
  });

  test("不正な page は 1 にフォールバックする", () => {
    const { result } = renderHook(() => useSearchQueryParams(), {
      wrapper: createWrapper("/app/ws1/search?page=-1"),
    });
    expect(result.current.query.page).toBe(1);
  });

  test("updateQuery は指定した項目だけを差し替える", () => {
    const { result } = renderHook(() => useSearchQueryParams(), {
      wrapper: createWrapper("/app/ws1/search?q=hello&filter=messages&page=3"),
    });

    act(() => {
      result.current.updateQuery({ page: 2 });
    });

    expect(result.current.query).toEqual({ q: "hello", filter: "messages", page: 2 });
  });
});
