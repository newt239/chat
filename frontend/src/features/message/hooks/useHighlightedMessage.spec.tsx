import { renderHook } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";

describe("useHighlightedMessage", () => {
  test("対象が無いときはハイライトしない", () => {
    const { result } = renderHook(() => useHighlightedMessage(true, null, () => true));

    expect(result.current).toBeNull();
  });

  test("対象のメッセージへスクロールしてハイライトする", () => {
    const calls: string[] = [];
    const { result } = renderHook(() =>
      useHighlightedMessage(true, "m1", (id) => {
        calls.push(id);
        return true;
      }),
    );

    expect(result.current).toBe("m1");
    expect(calls).toEqual(["m1"]);
  });

  test("一覧が揃うまではスクロールしない", () => {
    const calls: string[] = [];
    renderHook(() =>
      useHighlightedMessage(false, "m1", (id) => {
        calls.push(id);
        return true;
      }),
    );

    expect(calls).toEqual([]);
  });
});
