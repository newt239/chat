import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createStore, Provider } from "jotai";
import { expect, test, vi } from "vite-plus/test";

import { SidebarSection } from "./SidebarSection";

test("見出しで折りたたみ、追加ボタンで onAdd を呼ぶ", async () => {
  const onAdd = vi.fn<() => void>();
  render(
    <Provider store={createStore()}>
      <SidebarSection id="channels" title="チャンネル" onAdd={{ label: "作成", onPress: onAdd }}>
        <p>general</p>
      </SidebarSection>
    </Provider>,
  );

  const heading = screen.getByRole("button", { name: "チャンネル" });
  expect(heading).toHaveAttribute("aria-expanded", "true");
  expect(screen.getByText("general")).toBeInTheDocument();

  await userEvent.click(screen.getByRole("button", { name: "作成" }));
  expect(onAdd).toHaveBeenCalledOnce();

  await userEvent.click(heading);
  expect(heading).toHaveAttribute("aria-expanded", "false");
});
