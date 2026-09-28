import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageSchema, DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { CreateDMModal } from "./CreateDMModal";

import type { CreateChannelRequest } from "#/gen/chat/v1/channel_service_pb";
import type {
  CreateDirectMessageRequest,
  CreateGroupDirectMessageRequest,
} from "#/gen/chat/v1/direct_message_service_pb";

const userId = (index: number) =>
  `00000000-0000-0000-0000-0000000001${String(index).padStart(2, "0")}`;
const others = Array.from({ length: 11 }, (_, index) => ({
  displayName: `User ${String(index + 1).padStart(2, "0")}`,
  email: `user${index + 1}@example.com`,
  userId: userId(index + 1),
}));
const members = [
  { displayName: currentUser.displayName, email: currentUser.email, userId: currentUser.id },
  ...others,
].map((member) => create(WorkspaceMemberSchema, member));

const setup = async () => {
  const createDM = vi.fn<(req: CreateDirectMessageRequest) => void>();
  const createGroupDM = vi.fn<(req: CreateGroupDirectMessageRequest) => void>();
  const createChannel = vi.fn<(req: CreateChannelRequest) => void>();
  const onClose = vi.fn<() => void>();
  const { router } = await renderWithProviders(
    <CreateDMModal workspaceId="ws1" opened onClose={onClose} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(WorkspaceService.method.listMembers, () => ({ members }));
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [create(ChannelSchema, { id: "c1", name: "general" })],
      }));
      routes.rpc(DirectMessageService.method.createDirectMessage, (req) => {
        createDM(req);
        return { directMessage: create(DirectMessageSchema, { id: "dm1" }) };
      });
      routes.rpc(DirectMessageService.method.createGroupDirectMessage, (req) => {
        createGroupDM(req);
        return { directMessage: create(DirectMessageSchema, { id: "gdm1" }) };
      });
      routes.rpc(ChannelService.method.createChannel, (req) => {
        createChannel(req);
        return { channel: create(ChannelSchema, { id: "c2", name: req.name }) };
      });
    },
  );
  await screen.findByRole("option", { name: "User 01" });
  return { createChannel, createDM, createGroupDM, onClose, router };
};

const pick = async (name: string) => {
  await userEvent.click(screen.getByRole("option", { name }));
};

describe("CreateDMModal", () => {
  test("自分以外のメンバーを候補に出し、検索で絞り込める", async () => {
    await setup();
    expect(screen.queryByRole("option", { name: "Alice" })).not.toBeInTheDocument();
    expect(screen.getAllByRole("option")).toHaveLength(11);

    await userEvent.type(screen.getByRole("searchbox", { name: "メンバーを検索" }), "user 02");
    expect(screen.getAllByRole("option")).toHaveLength(1);
    expect(screen.getByRole("option", { name: "User 02" })).toBeInTheDocument();
  });

  test("選んだ相手をチップで表示し、1 人なら DM を作って開く", async () => {
    const { createDM, router } = await setup();
    expect(screen.getByText(/あなたを含めて 1 \/ 10 人/)).toBeInTheDocument();

    await pick("User 01");
    expect(screen.getByText(/あなたを含めて 2 \/ 10 人/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "User 01 を外す" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "DM を開始" }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/dm1");
    });
    expect(createDM).toHaveBeenCalledWith(expect.objectContaining({ userId: userId(1) }));
  });

  test("チップの削除ボタンで選択を外す", async () => {
    await setup();
    await pick("User 01");
    await userEvent.click(screen.getByRole("button", { name: "User 01 を外す" }));
    expect(screen.queryByRole("button", { name: "User 01 を外す" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "DM を開始" })).toBeDisabled();
  });

  test("2 人以上ならグループ DM を作る", async () => {
    const { createGroupDM } = await setup();
    await pick("User 01");
    await pick("User 02");
    await userEvent.click(screen.getByRole("button", { name: "グループ DM を開始" }));
    await waitFor(() => {
      expect(createGroupDM).toHaveBeenCalledWith(
        expect.objectContaining({ userIds: [userId(1), userId(2)], workspaceId: "ws1" }),
      );
    });
  });

  test("10 人を超えたらチャンネル名を入力させ、非公開チャンネルを作る", async () => {
    const { createChannel, router } = await setup();
    for (const member of others.slice(0, 10)) {
      await pick(member.displayName);
    }
    expect(screen.getByText(/あなたを含めて 11 \/ 10 人/)).toBeInTheDocument();
    expect(
      screen.getByText("10 人を超えるため、非公開チャンネルとして作成します。"),
    ).toBeInTheDocument();

    const submit = screen.getByRole("button", { name: "非公開チャンネルを作成" });
    await userEvent.click(submit);
    expect(await screen.findByText("チャンネル名を入力してください")).toBeInTheDocument();
    expect(createChannel).not.toHaveBeenCalled();

    const nameField = screen.getByRole("textbox", { name: "チャンネル名" });
    await userEvent.type(nameField, "General");
    expect(nameField).toHaveValue("general");
    expect(screen.getByText("#general はすでにあります")).toBeInTheDocument();

    await userEvent.clear(nameField);
    await userEvent.type(nameField, "war-room");
    await userEvent.click(submit);
    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/c2");
    });
    expect(createChannel).toHaveBeenCalledWith(
      expect.objectContaining({
        isPrivate: true,
        memberIds: others.slice(0, 10).map((member) => member.userId),
        name: "war-room",
        workspaceId: "ws1",
      }),
    );
  });
});
