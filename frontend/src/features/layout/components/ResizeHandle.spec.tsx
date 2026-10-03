import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vite-plus/test";

import { ResizeHandle } from "./ResizeHandle";

const renderHandle = (edge: "left" | "right") => {
  const onChange = vi.fn<(value: number) => void>();
  render(
    <ResizeHandle
      label="サイドバーの幅"
      edge={edge}
      value={300}
      minValue={200}
      maxValue={310}
      defaultValue={248}
      onChange={onChange}
    />,
  );
  return { handle: screen.getByRole("separator", { name: "サイドバーの幅" }), onChange };
};

test("矢印キーで幅を変え、上限で止める", () => {
  const { handle, onChange } = renderHandle("right");
  expect(handle).toHaveAttribute("aria-valuenow", "300");

  fireEvent.keyDown(handle, { key: "ArrowLeft" });
  fireEvent.keyDown(handle, { key: "ArrowRight" });

  expect(onChange.mock.calls).toStrictEqual([[284], [310]]);
});

test("左端のつまみは左へドラッグすると広がる", () => {
  const { handle, onChange } = renderHandle("left");
  handle.setPointerCapture = vi.fn<(pointerId: number) => void>();

  fireEvent.pointerDown(handle, { clientX: 500, pointerId: 1 });
  fireEvent.pointerMove(handle, { clientX: 450, pointerId: 1 });

  expect(onChange).toHaveBeenLastCalledWith(310);
});

test("ダブルクリックで既定の幅に戻す", () => {
  const { handle, onChange } = renderHandle("right");

  fireEvent.doubleClick(handle);

  expect(onChange).toHaveBeenCalledWith(248);
});
