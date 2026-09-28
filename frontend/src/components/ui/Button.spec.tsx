import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Button } from "./Button";

describe("Button", () => {
  test("押すと onPress が呼ばれる", async () => {
    const onPress = vi.fn<() => void>();
    render(<Button onPress={onPress}>保存</Button>);

    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(onPress).toHaveBeenCalledOnce();
  });

  test("無効なときは押せない", async () => {
    const onPress = vi.fn<() => void>();
    render(
      <Button isDisabled onPress={onPress}>
        保存
      </Button>,
    );

    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(onPress).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });

  test("処理中は押せず、状態を data 属性で表す", async () => {
    const onPress = vi.fn<() => void>();
    render(
      <Button isPending onPress={onPress}>
        保存
      </Button>,
    );

    const button = screen.getByRole("button", { name: /保存/ });
    await userEvent.click(button);

    expect(onPress).not.toHaveBeenCalled();
    expect(button).toHaveAttribute("data-pending", "true");
  });

  test("利用側の className で既定のクラスを上書きできる", () => {
    render(
      <Button variant="secondary" className="h-10">
        保存
      </Button>,
    );

    const button = screen.getByRole("button", { name: "保存" });
    expect(button).toHaveClass("h-10", "bg-surface");
    expect(button).not.toHaveClass("h-8");
  });
});
