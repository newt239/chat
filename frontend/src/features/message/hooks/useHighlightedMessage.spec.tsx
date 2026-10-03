import { renderHook } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";

describe("useHighlightedMessage", () => {
  test("対象が無いときはハイライトしない", () => {
    const { result } = renderHook(() => useHighlightedMessage(null, () => true));

    expect(result.current).toBeNull();
  });

  test("対象のメッセージへスクロールしてハイライトする", () => {
    const calls: string[] = [];
    const { result } = renderHook(() =>
      useHighlightedMessage("m1", (id) => {
        calls.push(id);
        return true;
      }),
    );

    expect(result.current).toBe("m1");
    expect(calls).toEqual(["m1"]);
  });

  test("対象が一覧に無ければ、作り直された関数で再び試す", () => {
    const calls: string[] = [];
    const { rerender } = renderHook(
      ({ found }: { found: boolean }) =>
        useHighlightedMessage("m1", (id) => {
          calls.push(id);
          return found;
        }),
      { initialProps: { found: false } },
    );
    rerender({ found: true });
    rerender({ found: true });

    expect(calls).toEqual(["m1", "m1"]);
  });
});
