import { useState } from "react";

import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DialogTrigger } from "react-aria-components";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { Button } from "#/components/ui/Button/Button";

import { ResponsivePopover } from "./ResponsivePopover";

const stubMatchMedia = (matches: boolean) => {
  vi.stubGlobal("matchMedia", (query: string) => ({
    addEventListener: () => {},
    matches,
    media: query,
    removeEventListener: () => {},
  }));
};

const Example = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <DialogTrigger isOpen={isOpen} onOpenChange={setIsOpen}>
      <Button>開く</Button>
      <ResponsivePopover aria-label="ピッカー" isOpen={isOpen} onOpenChange={setIsOpen}>
        <p>中身</p>
      </ResponsivePopover>
    </DialogTrigger>
  );
};

describe("ResponsivePopover", () => {
  afterEach(() => {
    stubMatchMedia(false);
  });

  test("デスクトップではポップオーバーで開く", async () => {
    render(<Example />);
    await userEvent.click(screen.getByRole("button", { name: "開く" }));
    expect(await screen.findByRole("dialog", { name: "ピッカー" })).toHaveTextContent("中身");
  });

  test("モバイルでは見出し付きのシートで開き、Escape で閉じる", async () => {
    stubMatchMedia(true);
    render(<Example />);
    await userEvent.click(screen.getByRole("button", { name: "開く" }));
    const sheet = await screen.findByRole("dialog", { name: "ピッカー" });
    expect(sheet).toHaveTextContent("中身");
    await userEvent.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });
});
