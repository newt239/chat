import type { ReactNode } from "react";

import { renderHook } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, test } from "vite-plus/test";

import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";

const createWrapper = (initialPath: string) => {
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  );
  return Wrapper;
};

describe("useHighlightedMessage", () => {
  test("message クエリが無いときは対象を返さない", () => {
    const { result } = renderHook(() => useHighlightedMessage(true), {
      wrapper: createWrapper("/app/ws1/ch1"),
    });

    expect(result.current.targetMessageId).toBeNull();
    expect(result.current.highlightedId).toBeNull();
  });

  test("message クエリのメッセージをハイライト対象にする", () => {
    const { result } = renderHook(() => useHighlightedMessage(true), {
      wrapper: createWrapper("/app/ws1/ch1?message=m1"),
    });

    expect(result.current.targetMessageId).toBe("m1");
    expect(result.current.highlightedId).toBe("m1");
  });
});
