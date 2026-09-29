import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { Tab } from "#/components/ui/Tab/Tab";
import { TabList } from "#/components/ui/TabList/TabList";
import { TabPanel } from "#/components/ui/TabPanel/TabPanel";

import { Tabs } from "./Tabs";

import type { Key } from "react-aria-components";

const renderTabs = (onSelectionChange = vi.fn<(key: Key) => void>()) =>
  render(
    <Tabs onSelectionChange={onSelectionChange}>
      <TabList aria-label="表示">
        <Tab id="messages">メッセージ</Tab>
        <Tab id="files">ファイル</Tab>
      </TabList>
      <TabPanel id="messages">メッセージ一覧</TabPanel>
      <TabPanel id="files">ファイル一覧</TabPanel>
    </Tabs>,
  );

describe("Tabs", () => {
  test("選んだタブのパネルだけを表示する", async () => {
    const onSelectionChange = vi.fn<(key: Key) => void>();
    renderTabs(onSelectionChange);
    expect(screen.getByRole("tabpanel")).toHaveTextContent("メッセージ一覧");

    await userEvent.click(screen.getByRole("tab", { name: "ファイル" }));

    expect(screen.getByRole("tab", { name: "ファイル" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tabpanel")).toHaveTextContent("ファイル一覧");
    expect(onSelectionChange).toHaveBeenLastCalledWith("files");
  });

  test("矢印キーでタブを移動できる", async () => {
    renderTabs();

    await userEvent.tab();
    await userEvent.keyboard("{ArrowRight}");

    expect(screen.getByRole("tab", { name: "ファイル" })).toHaveAttribute("aria-selected", "true");
  });
});
