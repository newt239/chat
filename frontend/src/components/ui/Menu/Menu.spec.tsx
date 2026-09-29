import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSection } from "#/components/ui/MenuSection/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";

import { Menu } from "./Menu";

describe("Menu", () => {
  test("トリガーで開き、選んだ項目の操作を実行して閉じる", async () => {
    const onPin = vi.fn<() => void>();
    const onDelete = vi.fn<() => void>();
    render(
      <Menu
        trigger={
          <IconButton label="その他">
            <svg />
          </IconButton>
        }
      >
        <MenuSection title="操作">
          <MenuItem shortcut="P" onAction={onPin}>
            ピン留め
          </MenuItem>
        </MenuSection>
        <MenuSeparator />
        <MenuItem tone="danger" onAction={onDelete}>
          削除
        </MenuItem>
      </Menu>,
    );

    await userEvent.click(screen.getByRole("button", { name: "その他" }));
    const menu = await screen.findByRole("menu", { name: "その他" });
    expect(menu).toHaveTextContent("操作");
    expect(screen.getByRole("separator")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("menuitem", { name: /ピン留め/ }));

    expect(onPin).toHaveBeenCalledOnce();
    expect(onDelete).not.toHaveBeenCalled();
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });
});
