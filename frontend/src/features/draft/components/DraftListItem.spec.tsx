import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DraftSchema } from "#/gen/chat/v1/draft_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { DraftListItem } from "./DraftListItem";

const render = (parentId: string | undefined, onDelete: () => void) =>
  renderWithProviders(
    <DraftListItem
      workspaceId="ws1"
      draft={create(DraftSchema, { body: "書きかけの本文", channelId: "c1", id: "d1", parentId })}
      label="#general"
      onDelete={onDelete}
    />,
    "/app/ws1/drafts",
    () => {},
  );

describe("DraftListItem", () => {
  test("会話の名前と本文を出し、チャンネルを開くリンクと削除を置く", async () => {
    const onDelete = vi.fn<() => void>();
    await render(undefined, onDelete);

    expect(screen.getByText("#general")).toBeInTheDocument();
    expect(screen.getByText("書きかけの本文")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "開く" })).toHaveAttribute("href", "/app/ws1/c1");
    await userEvent.click(screen.getByRole("button", { name: "下書きを削除" }));
    expect(onDelete).toHaveBeenCalled();
  });

  test("スレッドの下書きはスレッドを開く", async () => {
    await render("m1", vi.fn<() => void>());

    expect(screen.getByText("スレッドへの返信")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "開く" })).toHaveAttribute(
      "href",
      "/app/ws1/c1/thread/m1",
    );
  });
});
