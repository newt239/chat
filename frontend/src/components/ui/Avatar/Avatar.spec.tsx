import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ColorModeContext } from "#/providers/theme/colorMode";

import { Avatar } from "./Avatar";

describe("Avatar", () => {
  test("画像がなければ名前の頭文字を表示する", () => {
    render(<Avatar name="misaki" />);

    const avatar = screen.getByRole("img", { name: "misaki" });
    expect(avatar).toHaveTextContent("M");
    expect(avatar).toHaveStyle({ height: "32px", width: "32px" });
  });

  test("画像があれば画像を表示する", () => {
    const { container } = render(<Avatar name="Kenta" src="https://example.com/a.png" />);

    expect(container.querySelector("img")).toHaveAttribute("src", "https://example.com/a.png");
  });

  test("在席状態の点を表示する", () => {
    const { container } = render(<Avatar name="Ren" presence="online" />);

    expect(container.querySelector("[data-presence]")).toHaveAttribute("data-presence", "online");
  });

  test("同じ名前でも表示モードで背景色を変える", () => {
    render(
      <>
        <Avatar name="Yui" />
        <ColorModeContext value="dark">
          <Avatar name="Yui" />
        </ColorModeContext>
      </>,
    );

    const [light, dark] = screen.getAllByRole("img", { name: "Yui" });
    expect(light?.style.backgroundColor).not.toBe(dark?.style.backgroundColor);
  });
});
