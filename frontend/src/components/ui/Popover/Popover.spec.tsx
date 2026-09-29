import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DialogTrigger } from "react-aria-components";
import { describe, expect, test } from "vite-plus/test";

import { Button } from "#/components/ui/Button/Button";

import { Popover } from "./Popover";

describe("Popover", () => {
  test("トリガーで開き、Escape で閉じる", async () => {
    render(
      <DialogTrigger>
        <Button>プロフィール</Button>
        <Popover aria-label="プロフィールの詳細">田中 美咲</Popover>
      </DialogTrigger>,
    );

    await userEvent.click(screen.getByRole("button", { name: "プロフィール" }));
    expect(await screen.findByRole("dialog", { name: "プロフィールの詳細" })).toHaveTextContent(
      "田中 美咲",
    );

    await userEvent.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
