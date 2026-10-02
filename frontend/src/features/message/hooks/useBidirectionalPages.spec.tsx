import type { ReactNode } from "react";

import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { useBidirectionalPages } from "./useBidirectionalPages";

import type { PageCursor } from "./useBidirectionalPages";

import type { Timestamp } from "@bufbuild/protobuf/wkt";

type Item = { id: number; createdAt: Timestamp };
type Page = { items: Item[]; hasMore: boolean; hasNewer: boolean };

const item = (id: number) => ({ createdAt: timestampFromMs(id * 1000), id });
const getItems = (page: Page) => page.items;

// 1〜9 の項目のうち、最初は 4〜6 を返す。前後は 3 件ずつ返す
const fetchPage = (cursor: PageCursor) => {
  if (cursor === null) {
    return Promise.resolve({ hasMore: true, hasNewer: true, items: [item(4), item(5), item(6)] });
  }
  const boundary = Number(cursor.boundary.seconds);
  return Promise.resolve(
    cursor.direction === "older"
      ? {
          hasMore: false,
          hasNewer: true,
          items: [1, 2, 3].filter((id) => id < boundary).map((id) => item(id)),
        }
      : {
          hasMore: true,
          hasNewer: false,
          items: [7, 8, 9].filter((id) => id > boundary).map((id) => item(id)),
        },
  );
};

const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
);

describe("useBidirectionalPages", () => {
  test("古い順の並びでは、古い側を先頭に・新しい側を末尾に足す", async () => {
    const fetchSpy = vi.fn(fetchPage);
    const { result } = renderHook(
      () =>
        useBidirectionalPages({
          enabled: true,
          fetchPage: fetchSpy,
          getItems,
          newestFirst: false,
          queryKey: ["pages"],
        }),
      { wrapper },
    );
    await waitFor(() => {
      expect(result.current.items?.map(({ id }) => id)).toEqual([4, 5, 6]);
    });
    expect(result.current.hasOlder).toBe(true);
    expect(result.current.hasNewer).toBe(true);

    act(() => {
      result.current.load("older");
    });
    await waitFor(() => {
      expect(result.current.items?.map(({ id }) => id)).toEqual([1, 2, 3, 4, 5, 6]);
    });
    expect(result.current.hasOlder).toBe(false);

    act(() => {
      result.current.load("newer");
    });
    await waitFor(() => {
      expect(result.current.items?.map(({ id }) => id)).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9]);
    });
    expect(result.current.hasNewer).toBe(false);
    expect(fetchSpy).toHaveBeenCalledTimes(3);
  });

  test("無効なあいだは取得しない", () => {
    const fetchSpy = vi.fn(fetchPage);
    const { result } = renderHook(
      () =>
        useBidirectionalPages({
          enabled: false,
          fetchPage: fetchSpy,
          getItems,
          newestFirst: true,
          queryKey: ["disabled"],
        }),
      { wrapper },
    );
    expect(result.current.items).toBeUndefined();
    expect(fetchSpy).not.toHaveBeenCalled();
  });
});
