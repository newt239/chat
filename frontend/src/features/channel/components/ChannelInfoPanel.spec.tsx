import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelInfoPanel } from "./ChannelInfoPanel";

import type { UpdateChannelRequest } from "#/gen/chat/v1/channel_service_pb";

const setup = async () => {
  const updateChannel = vi.fn<(req: UpdateChannelRequest) => void>();
  await renderWithProviders(
    <ChannelInfoPanel workspaceId="ws1" channelId="c1" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [
          create(ChannelSchema, { description: "フロントの話題", id: "c1", name: "dev/frontend" }),
        ],
      }));
      routes.rpc(ChannelService.method.updateChannel, (req) => {
        updateChannel(req);
        return {};
      });
    },
  );
  await screen.findByRole("heading", { name: "dev/frontend" });
  return { updateChannel };
};

describe("ChannelInfoPanel", () => {
  test("チャンネルの概要を表示し、メンバーの管理は出さない", async () => {
    await setup();
    expect(screen.getByText("フロントの話題", { selector: "p" })).toBeInTheDocument();
    expect(screen.getByText("公開")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "退出する" })).not.toBeInTheDocument();
  });

  test("参加していないチャンネルも個別に取得して表示する", async () => {
    await renderWithProviders(
      <ChannelInfoPanel workspaceId="ws1" channelId="c2" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelService.method.listChannels, () => ({ channels: [] }));
        routes.rpc(ChannelService.method.getChannel, () => ({
          channel: create(ChannelSchema, { description: "雑談", id: "c2", name: "random" }),
        }));
      },
    );
    expect(await screen.findByRole("heading", { name: "random" })).toBeInTheDocument();
    expect(screen.getByText("雑談", { selector: "p" })).toBeInTheDocument();
  });

  test("設定では末尾の名前だけを編集し、親のパスを付けて保存する", async () => {
    const { updateChannel } = await setup();
    const nameField = screen.getByRole("textbox", { name: "名前" });
    expect(nameField).toHaveValue("frontend");

    await userEvent.clear(nameField);
    await userEvent.type(nameField, "web");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(updateChannel).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", name: "dev/web" }),
      );
    });
  });
});
