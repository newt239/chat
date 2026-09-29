import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { IconButton } from "./IconButton";

describe("IconButton", () => {
  test("label をアクセシブルな名前にする", async () => {
    const onPress = vi.fn<() => void>();
    render(
      <IconButton label="閉じる" onPress={onPress}>
        <svg />
      </IconButton>,
    );

    await userEvent.click(screen.getByRole("button", { name: "閉じる" }));

    expect(onPress).toHaveBeenCalledOnce();
  });

  test("キーボードでフォーカスするとツールチップを表示する", async () => {
    render(
      <IconButton label="閉じる">
        <svg />
      </IconButton>,
    );

    await userEvent.tab();

    expect(await screen.findByRole("tooltip")).toHaveTextContent("閉じる");
  });
});
