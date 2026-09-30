import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Button } from "react-aria-components";
import { describe, expect, test, vi } from "vite-plus/test";

import { Menu } from "../Menu/Menu";
import { MenuItem } from "../MenuItem/MenuItem";
import { Submenu } from "./Submenu";

describe("Submenu", () => {
  test("項目を押すと入れ子のメニューを開き、選んだ項目を実行する", async () => {
    const onAction = vi.fn<() => void>();
    render(
      <Menu trigger={<Button>開く</Button>}>
        <Submenu label="移動" icon={null}>
          <MenuItem onAction={onAction}>仕事</MenuItem>
        </Submenu>
      </Menu>,
    );

    await userEvent.click(screen.getByRole("button", { name: "開く" }));
    const trigger = screen.getByRole("menuitem", { name: "移動" });
    expect(trigger).toHaveAttribute("aria-haspopup", "menu");
    await userEvent.click(trigger);
    await userEvent.click(await screen.findByRole("menuitem", { name: "仕事" }));
    expect(onAction).toHaveBeenCalledOnce();
  });
});
