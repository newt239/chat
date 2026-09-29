import { act, renderHook } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vite-plus/test";

import { useVisualViewport } from "./useVisualViewport";

// setup で差し替えた localStorage なども戻ってしまうため unstubAllGlobals は使わない
afterEach(() => {
  vi.stubGlobal("visualViewport", undefined);
});

test("キーボードで見える範囲が縮むと高さを合わせて開いているとみなす", () => {
  const viewport = Object.assign(new EventTarget(), { height: 800 });
  vi.stubGlobal("visualViewport", viewport);
  vi.stubGlobal("innerHeight", 800);
  vi.stubGlobal("scrollTo", vi.fn());

  const { result } = renderHook(() => useVisualViewport());
  expect(result.current).toEqual({ height: 800, keyboardOpen: false });

  viewport.height = 450;
  act(() => {
    viewport.dispatchEvent(new Event("resize"));
  });
  expect(result.current).toEqual({ height: 450, keyboardOpen: true });
});

test("visualViewport がなければ何もしない", () => {
  vi.stubGlobal("visualViewport", undefined);
  const { result } = renderHook(() => useVisualViewport());
  expect(result.current).toBeUndefined();
});
