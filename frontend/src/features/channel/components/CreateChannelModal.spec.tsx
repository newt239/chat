import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { CreateChannelModal } from "./CreateChannelModal";

import type { CreateChannelRequest } from "#/gen/chat/v1/channel_service_pb";

const setup = async (parentId: string | null = null) => {
  const createChannel = vi.fn<(req: CreateChannelRequest) => void>();
  const onClose = vi.fn<() => void>();
  const { router } = await renderWithProviders(
    <CreateChannelModal workspaceId="ws1" parentId={parentId} opened onClose={onClose} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [
          create(ChannelSchema, { id: "c1", name: "dev" }),
          create(ChannelSchema, { id: "c3", isPrivate: true, name: "secret" }),
        ],
      }));
      routes.rpc(ChannelService.method.createChannel, (req) => {
        createChannel(req);
        return { channel: create(ChannelSchema, { id: "c2", name: req.name }) };
      });
    },
  );
  const nameField = screen.getByRole("textbox", { name: "チャンネル名" });
  return { createChannel, nameField, onClose, router };
};

describe("CreateChannelModal", () => {
  test("親を指定して開くと親のパスと公開範囲を初期値にする", async () => {
    const { nameField } = await setup("c3");
    await waitFor(() => {
      expect(nameField).toHaveValue("secret/");
    });
    expect(screen.getByRole("switch", { name: /非公開/ })).toBeChecked();
  });

  test("スラッシュ区切りの名前から作成される階層と、新しく作られる親を表示する", async () => {
    const { nameField } = await setup();
    await userEvent.type(nameField, "Dev/Frontend/Web");
    expect(nameField).toHaveValue("dev/frontend/web");

    const preview = screen.getByRole("list", { name: "作成される場所" });
    expect(
      within(preview)
        .getAllByRole("listitem")
        .map((item) => item.textContent),
    ).toEqual(["# dev", "# frontend", "# web"]);
    expect(await screen.findByText("親の #dev/frontend も作成されます")).toBeInTheDocument();
  });

  test.each([
    ["dev", "#dev はすでにあります"],
    ["dev//web", "スラッシュの前後には名前が必要です"],
    ["a/b/c/d/e", "階層は 4 段までにしてください"],
    ["dev.web", "小文字の英数字・ハイフン・アンダースコアだけが使えます"],
    ["a".repeat(33), "各階層は 32 文字以内にしてください"],
  ])("%s は作成せずにエラーを表示する", async (name, message) => {
    const { createChannel, nameField } = await setup();
    await screen.findByText("「/」で区切ると親の下にツリー表示されます");
    await userEvent.type(nameField, name);
    await userEvent.click(screen.getByRole("button", { name: "作成" }));
    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(createChannel).not.toHaveBeenCalled();
  });

  test("作成したら閉じてチャンネルを開く", async () => {
    const { createChannel, nameField, onClose, router } = await setup();
    await userEvent.type(nameField, "dev/frontend");
    await userEvent.click(screen.getByRole("switch", { name: "非公開にする" }));
    await userEvent.click(screen.getByRole("button", { name: "作成" }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/c2");
    });
    expect(onClose).toHaveBeenCalled();
    expect(createChannel).toHaveBeenCalledWith(
      expect.objectContaining({ isPrivate: true, name: "dev/frontend", workspaceId: "ws1" }),
    );
  });
});
