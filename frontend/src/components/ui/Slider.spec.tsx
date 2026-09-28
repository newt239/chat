import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test, vi } from "vite-plus/test";

import { Slider } from "./Slider";

test("キーボードで値を動かし、変更中と確定の両方を通知する", () => {
  const onChange = vi.fn<(value: number) => void>();
  const onChangeEnd = vi.fn<(value: number) => void>();
  render(
    <Slider
      label="色相"
      value={10}
      minValue={0}
      maxValue={359}
      step={1}
      onChange={onChange}
      onChangeEnd={onChangeEnd}
    />,
  );

  const slider = screen.getByRole("slider", { name: "色相" });
  expect(slider).toHaveValue("10");
  fireEvent.keyDown(slider, { key: "ArrowRight" });
  fireEvent.keyUp(slider, { key: "ArrowRight" });

  expect(onChange).toHaveBeenCalledWith(11);
  expect(onChangeEnd).toHaveBeenCalledWith(11);
});
