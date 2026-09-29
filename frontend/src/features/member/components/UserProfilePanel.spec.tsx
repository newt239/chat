import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DirectMessageSchema, DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { UserNoteSchema, UserService } from "#/gen/chat/v1/user_service_pb";
import {
  WorkspaceMemberSchema,
  WorkspaceRole,
  WorkspaceService,
} from "#/gen/chat/v1/workspace_service_pb";
import { currentUser, renderWithProviders } from "#/test/renderWithProviders";

import { UserProfilePanel } from "./UserProfilePanel";

import type { UpdateUserNoteRequest } from "#/gen/chat/v1/user_service_pb";

const updateNote = vi.fn<(req: UpdateUserNoteRequest) => void>();

const render = (userId: string) =>
  renderWithProviders(
    <UserProfilePanel workspaceId="ws1" userId={userId} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [
          create(WorkspaceMemberSchema, {
            displayName: currentUser.displayName,
            email: currentUser.email,
            role: WorkspaceRole.OWNER,
            userId: currentUser.id,
          }),
          create(WorkspaceMemberSchema, {
            bio: "フロントエンド担当",
            displayName: "Bob",
            email: "bob@example.com",
            role: WorkspaceRole.ADMIN,
            timezone: "America/New_York",
            userId: "u-bob",
          }),
        ],
      }));
      routes.rpc(UserService.method.getUserNote, () => ({
        note: create(UserNoteSchema, { memo: "朝型", targetUserId: "u-bob" }),
      }));
      routes.rpc(UserService.method.updateUserNote, (req) => {
        updateNote(req);
        return {};
      });
      routes.rpc(DirectMessageService.method.createDirectMessage, () => ({
        directMessage: create(DirectMessageSchema, { id: "dm1" }),
      }));
    },
  );

describe("UserProfilePanel", () => {
  test("タイムゾーンを設定した相手には現地時刻を表示する", async () => {
    await render("u-bob");
    expect(await screen.findByText("America/New_York")).toBeInTheDocument();
    expect(screen.getByText("現地時刻")).toBeInTheDocument();
  });

  test("タイムゾーンが未設定なら現地時刻を表示しない", async () => {
    await render(currentUser.id);
    expect(
      await screen.findByRole("heading", { name: currentUser.displayName }),
    ).toBeInTheDocument();
    expect(screen.queryByText("現地時刻")).not.toBeInTheDocument();
  });

  test("プロフィールを表示し、メッセージボタンで DM を開く", async () => {
    const { router } = await render("u-bob");
    expect(await screen.findByRole("heading", { name: "Bob" })).toBeInTheDocument();
    expect(screen.getByText("管理者")).toBeInTheDocument();
    expect(screen.getByText("bob@example.com")).toBeInTheDocument();
    expect(screen.getByText("フロントエンド担当")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "メッセージ" }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe("/app/ws1/dm1");
    });
  });

  test("自分のプロフィールにはメッセージボタンを出さない", async () => {
    await render(currentUser.id);
    expect(await screen.findByText("オーナー")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "メッセージ" })).not.toBeInTheDocument();
  });

  test("自分だけのニックネームとメモを保存する", async () => {
    await render("u-bob");
    const memo = await screen.findByRole("textbox", { name: "メモ" });
    expect(memo).toHaveValue("朝型");
    await userEvent.type(screen.getByRole("textbox", { name: "表示名" }), "ボブさん");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(updateNote).toHaveBeenCalledWith(
        expect.objectContaining({ memo: "朝型", nickname: "ボブさん", targetUserId: "u-bob" }),
      );
    });
  });

  test("自分のプロフィールにはメモ欄を出さない", async () => {
    await render(currentUser.id);
    expect(await screen.findByText("オーナー")).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "メモ" })).not.toBeInTheDocument();
  });

  test("見つからないユーザーはその旨を表示する", async () => {
    await render("u-unknown");
    expect(await screen.findByText("ユーザーが見つかりませんでした")).toBeInTheDocument();
  });
});
