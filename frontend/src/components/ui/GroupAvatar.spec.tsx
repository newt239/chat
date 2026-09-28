import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { GroupAvatar } from "./GroupAvatar";

describe("GroupAvatar", () => {
  test("人数を表示し、読み上げ用の名前を付ける", () => {
    render(<GroupAvatar count={4} size={24} />);

    const avatar = screen.getByRole("img", { name: "4 人のグループ" });
    expect(avatar).toHaveTextContent("4");
    expect(avatar).toHaveStyle({ width: "24px" });
  });
});
