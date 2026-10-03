import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { channelViewersAtom } from "#/features/channel/atoms";
import {
  ChannelMemberSchema,
  ChannelMemberService,
  ChannelRole,
} from "#/gen/chat/v1/channel_member_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { ChannelMemberPanel } from "./ChannelMemberPanel";

import type {
  InviteChannelMemberRequest,
  LeaveChannelRequest,
  RemoveChannelMemberRequest,
  UpdateChannelMemberRoleRequest,
} from "#/gen/chat/v1/channel_member_service_pb";

const bob = { displayName: "Bob", email: "bob@example.com", userId: "u-bob" };
const carol = { displayName: "Carol", email: "carol@example.com", userId: "u-carol" };

const setupManaged = async () => {
  const updateRole = vi.fn<(req: UpdateChannelMemberRoleRequest) => void>();
  const invite = vi.fn<(req: InviteChannelMemberRequest) => void>();
  const remove = vi.fn<(req: RemoveChannelMemberRequest) => void>();
  const leave = vi.fn<(req: LeaveChannelRequest) => void>();
  await renderWithProviders(
    <ChannelMemberPanel workspaceId="ws1" channelId="c1" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
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
      routes.rpc(ChannelMemberService.method.removeChannelMember, (req) => {
        remove(req);
        return {};
      });
      routes.rpc(ChannelMemberService.method.leaveChannel, (req) => {
        leave(req);
        return {};
      });
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [bob, carol].map((member) => create(WorkspaceMemberSchema, member)),
      }));
    },
  );
  await screen.findByRole("button", { name: "Bob の操作" });
  return { invite, leave, remove, updateRole };
};

describe("ChannelMemberPanel", () => {
  test("メンバーを一覧し、押すとプロフィールを右パネルに開く", async () => {
    const { router } = await renderWithProviders(
      <ChannelMemberPanel workspaceId="ws1" channelId="c1" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({
          members: [
            create(ChannelMemberSchema, {
              displayName: "Bob",
              email: "bob@example.com",
              userId: "u-bob",
            }),
          ],
        }));
      },
    );

    const row = await screen.findByRole("button", { name: /Bob/ });
    expect(row).toHaveTextContent("bob@example.com");
    await userEvent.click(row);
    expect(router.state.location.search).toEqual({ profile: "u-bob" });
  });

  test("閲覧中のメンバーを分けて、ニックネームで表示する", async () => {
    const { store } = await renderWithProviders(
      <ChannelMemberPanel workspaceId="ws1" channelId="c1" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({
          members: [
            create(ChannelMemberSchema, { displayName: "Bob", email: "b@x", userId: "u-bob" }),
            create(ChannelMemberSchema, { displayName: "Carol", email: "c@x", userId: "u-carol" }),
          ],
        }));
        routes.rpc(WorkspaceService.method.listMembers, () => ({
          members: [create(WorkspaceMemberSchema, { nickname: "ボブさん", userId: "u-bob" })],
        }));
      },
    );
    store.set(channelViewersAtom, { c1: ["u-bob"] });

    const viewing = await screen.findByRole("region", { name: "いま閲覧中" });
    expect(await within(viewing).findByText("ボブさん")).toBeInTheDocument();
    expect(within(viewing).getByText("閲覧中")).toBeInTheDocument();
    const others = screen.getByRole("region", { name: "その他のメンバー" });
    expect(within(others).getByText("Carol")).toBeInTheDocument();
  });

  test("メンバーがいなければその旨を表示する", async () => {
    await renderWithProviders(
      <ChannelMemberPanel workspaceId="ws1" channelId="c1" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({ members: [] }));
      },
    );
    expect(await screen.findByText("まだメンバーがいません")).toBeInTheDocument();
  });

  test("メンバーのメニューからロールを変更する", async () => {
    const { updateRole } = await setupManaged();
    await userEvent.click(screen.getByRole("button", { name: "Bob の操作" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "管理者" }));
    await waitFor(() => {
      expect(updateRole).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", role: ChannelRole.ADMIN, userId: "u-bob" }),
      );
    });
  });

  test("メンバーを外す前に確認する", async () => {
    const { remove } = await setupManaged();
    await userEvent.click(screen.getByRole("button", { name: "Bob の操作" }));
    await userEvent.click(screen.getByRole("menuitem", { name: "チャンネルから外す" }));
    const dialog = await screen.findByRole("alertdialog");
    expect(remove).not.toHaveBeenCalled();
    await userEvent.click(within(dialog).getByRole("button", { name: "チャンネルから外す" }));
    await waitFor(() => {
      expect(remove).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", userId: "u-bob" }),
      );
    });
  });

  test("退出する前に確認する", async () => {
    const { leave } = await setupManaged();
    await userEvent.click(screen.getByRole("button", { name: "退出する" }));
    const dialog = await screen.findByRole("alertdialog");
    expect(leave).not.toHaveBeenCalled();
    await userEvent.click(within(dialog).getByRole("button", { name: "退出する" }));
    await waitFor(() => {
      expect(leave).toHaveBeenCalledWith(expect.objectContaining({ channelId: "c1" }));
    });
  });

  test("参加していないメンバーを招待できる", async () => {
    const { invite } = await setupManaged();
    await userEvent.click(screen.getByRole("combobox", { name: "メンバーを招待" }));
    await userEvent.click(await screen.findByRole("option", { name: "Carol" }));
    await userEvent.click(screen.getByRole("button", { name: "招待" }));
    await waitFor(() => {
      expect(invite).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "c1", userId: "u-carol" }),
      );
    });
  });
});
