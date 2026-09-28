import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WorkspaceDialogs } from "#/features/layout/components/WorkspaceDialogs";
import {
  UserGroupMemberSchema,
  UserGroupSchema,
  UserGroupService,
} from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupPanel } from "./UserGroupPanel";

import type {
  AddUserGroupMemberRequest,
  DeleteUserGroupRequest,
  RemoveUserGroupMemberRequest,
  UpdateUserGroupRequest,
} from "#/gen/chat/v1/user_group_service_pb";

const setup = async () => {
  const updateGroup = vi.fn<(req: UpdateUserGroupRequest) => void>();
  const deleteGroup = vi.fn<(req: DeleteUserGroupRequest) => void>();
  const addMember = vi.fn<(req: AddUserGroupMemberRequest) => void>();
  const removeMember = vi.fn<(req: RemoveUserGroupMemberRequest) => void>();
  await renderWithProviders(
    <>
      <UserGroupPanel workspaceId="ws1" groupId="g1" />
      <WorkspaceDialogs workspaceId="ws1" />
    </>,
    "/app/ws1?group=g1",
    (routes) => {
      routes.rpc(UserGroupService.method.listUserGroups, () => ({
        userGroups: [
          create(UserGroupSchema, { description: "フロント班", id: "g1", name: "frontend" }),
        ],
      }));
      routes.rpc(UserGroupService.method.listUserGroupMembers, () => ({
        members: [create(UserGroupMemberSchema, { userId: "u-bob" })],
      }));
      routes.rpc(UserGroupService.method.updateUserGroup, (req) => {
        updateGroup(req);
        return {};
      });
      routes.rpc(UserGroupService.method.deleteUserGroup, (req) => {
        deleteGroup(req);
        return {};
      });
      routes.rpc(UserGroupService.method.addUserGroupMember, (req) => {
        addMember(req);
        return {};
      });
      routes.rpc(UserGroupService.method.removeUserGroupMember, (req) => {
        removeMember(req);
        return {};
      });
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [
          create(WorkspaceMemberSchema, { displayName: "Bob", nickname: "ボブ", userId: "u-bob" }),
          create(WorkspaceMemberSchema, { displayName: "Carol", userId: "u-carol" }),
        ],
      }));
    },
  );
  await screen.findByRole("heading", { name: "@frontend" });
  return { addMember, deleteGroup, removeMember, updateGroup };
};

describe("UserGroupPanel", () => {
  test("メンバーをニックネームで表示し、追加と削除ができる", async () => {
    const { addMember, removeMember } = await setup();
    expect(screen.getByText("フロント班")).toBeInTheDocument();

    await userEvent.click(await screen.findByRole("button", { name: "ボブ をグループから外す" }));
    await userEvent.click(screen.getByRole("combobox", { name: "メンバーを追加" }));
    await userEvent.click(await screen.findByRole("option", { name: "Carol" }));
    await userEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(addMember).toHaveBeenCalledWith(
        expect.objectContaining({ groupId: "g1", userId: "u-carol" }),
      );
    });
    expect(removeMember).toHaveBeenCalledWith(
      expect.objectContaining({ groupId: "g1", userId: "u-bob" }),
    );
  });

  test("名前と説明を編集する", async () => {
    const { updateGroup } = await setup();
    await userEvent.click(screen.getByRole("link", { name: "編集" }));
    const name = await screen.findByRole("textbox", { name: "グループ名" });
    await userEvent.clear(name);
    await userEvent.type(name, "web");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => {
      expect(updateGroup).toHaveBeenCalledWith(
        expect.objectContaining({ description: "フロント班", groupId: "g1", name: "web" }),
      );
    });
  });

  test("確認してから削除する", async () => {
    const { deleteGroup } = await setup();
    await userEvent.click(screen.getByRole("button", { name: "削除" }));
    await userEvent.click(await screen.findByRole("button", { name: "削除" }));
    await waitFor(() => {
      expect(deleteGroup).toHaveBeenCalledWith(expect.objectContaining({ groupId: "g1" }));
    });
  });
});
