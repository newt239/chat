import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  ChannelMemberSchema,
  ChannelMemberService,
  ChannelRole,
} from "#/gen/chat/v1/channel_member_service_pb";
import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { ChannelInfoPanel } from "./ChannelInfoPanel";

import type {
  InviteChannelMemberRequest,
  UpdateChannelMemberRoleRequest,
} from "#/gen/chat/v1/channel_member_service_pb";
import type { UpdateChannelRequest } from "#/gen/chat/v1/channel_service_pb";

const bob = { displayName: "Bob", email: "bob@example.com", userId: "u-bob" };
const carol = { displayName: "Carol", email: "carol@example.com", userId: "u-carol" };

const setup = async () => {
  const updateChannel = vi.fn<(req: UpdateChannelRequest) => void>();
  const updateRole = vi.fn<(req: UpdateChannelMemberRoleRequest) => void>();
  const invite = vi.fn<(req: InviteChannelMemberRequest) => void>();
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
      routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({
        members: [
          create(ChannelMemberSchema, {
            displayName: currentUser.displayName,
            email: currentUser.email,
            role: ChannelRole.ADMIN,
            userId: currentUser.id,
          }),
          create(ChannelMemberSchema, { ...bob, role: ChannelRole.MEMBER }),
        ],
      }));
      routes.rpc(ChannelMemberService.method.updateChannelMemberRole, (req) => {
        updateRole(req);
        return {};
      });
      routes.rpc(ChannelMemberService.method.inviteChannelMember, (req) => {
        invite(req);
        return {};
      });
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [bob, carol].map((member) => create(WorkspaceMemberSchema, member)),
      }));
    },
  );
  await screen.findByRole("heading", { name: "dev/frontend" });
  return { invite, updateChannel, updateRole };
};

describe("ChannelInfoPanel", () => {
  test("チャンネルの概要とメンバーを表示する", async () => {
    await setup();
    expect(screen.getByText("フロントの話題", { selector: "p" })).toBeInTheDocument();
    expect(screen.getByText("公開")).toBeInTheDocument();
    expect(await screen.findByText("メンバー 2 人")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "退出する" })).toBeInTheDocument();
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
        routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({ members: [] }));
        routes.rpc(WorkspaceService.method.listMembers, () => ({ members: [] }));
      },
    );
    expect(await screen.findByRole("heading", { name: "random" })).toBeInTheDocument();
    expect(screen.getByText("雑談", { selector: "p" })).toBeInTheDocument();
  });

  test("メンバーのメニューからロールを変更する", async () => {
    const { updateRole } = await setup();
    await userEvent.click(await screen.findByRole("button", { name: "Bob の操作" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "管理者" }));
    await waitFor(() => {
      expect(updateRole).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", role: ChannelRole.ADMIN, userId: "u-bob" }),
      );
    });
  });

  test("参加していないメンバーを招待できる", async () => {
    const { invite } = await setup();
    await userEvent.click(await screen.findByRole("combobox", { name: "メンバーを招待" }));
    await userEvent.click(await screen.findByRole("option", { name: "Carol" }));
    await userEvent.click(screen.getByRole("button", { name: "招待" }));
    await waitFor(() => {
      expect(invite).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", userId: "u-carol" }),
      );
    });
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
