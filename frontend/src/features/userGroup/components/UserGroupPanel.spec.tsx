import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import {
  UserGroupMemberSchema,
  UserGroupSchema,
  UserGroupService,
} from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { UserGroupManager } from "./UserGroupManager";

import type {
  AddUserGroupMemberRequest,
  CreateUserGroupRequest,
  DeleteUserGroupRequest,
  RemoveUserGroupMemberRequest,
} from "#/gen/chat/v1/user_group_service_pb";

const setup = async () => {
  const createGroup = vi.fn<(req: CreateUserGroupRequest) => void>();
  const deleteGroup = vi.fn<(req: DeleteUserGroupRequest) => void>();
  const addMember = vi.fn<(req: AddUserGroupMemberRequest) => void>();
  const removeMember = vi.fn<(req: RemoveUserGroupMemberRequest) => void>();
  await renderWithProviders(<UserGroupManager workspaceId="ws1" />, "/app/ws1", (routes) => {
    routes.rpc(UserGroupService.method.listUserGroups, () => ({
      userGroups: [
        create(UserGroupSchema, { description: "フロント班", id: "g1", name: "frontend" }),
      ],
    }));
    routes.rpc(UserGroupService.method.listUserGroupMembers, () => ({
      members: [create(UserGroupMemberSchema, { userId: "u-bob" })],
    }));
    routes.rpc(UserGroupService.method.createUserGroup, (req) => {
      createGroup(req);
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
        create(WorkspaceMemberSchema, { displayName: "Bob", userId: "u-bob" }),
        create(WorkspaceMemberSchema, { displayName: "Carol", userId: "u-carol" }),
      ],
    }));
  });
  await screen.findByText("@frontend");
  return { addMember, createGroup, deleteGroup, removeMember };
};

describe("UserGroupManager", () => {
  test("グループとメンバーを表示し、削除・メンバーの追加と削除ができる", async () => {
    const { addMember, deleteGroup, removeMember } = await setup();
    expect(screen.getByText("フロント班")).toBeInTheDocument();

    await userEvent.click(await screen.findByRole("button", { name: "Bob をグループから外す" }));
    await userEvent.click(screen.getByRole("combobox", { name: "メンバーを追加" }));
    await userEvent.click(await screen.findByRole("option", { name: "Carol" }));
    await userEvent.click(screen.getByRole("button", { name: "追加" }));
    await userEvent.click(screen.getByRole("button", { name: "@frontend を削除" }));

    await waitFor(() => {
      expect(deleteGroup).toHaveBeenCalledWith(expect.objectContaining({ groupId: "g1" }));
    });
    expect(removeMember).toHaveBeenCalledWith(
      expect.objectContaining({ groupId: "g1", userId: "u-bob" }),
    );
    expect(addMember).toHaveBeenCalledWith(
      expect.objectContaining({ groupId: "g1", userId: "u-carol" }),
    );
  });

  test("名前を入力してグループを作成する", async () => {
    const { createGroup } = await setup();
    const createButton = screen.getByRole("button", { name: "作成" });
    expect(createButton).toBeDisabled();

    await userEvent.type(screen.getByRole("textbox", { name: "グループを作成" }), " design ");
    await userEvent.click(createButton);
    await waitFor(() => {
      expect(createGroup).toHaveBeenCalledWith(
        expect.objectContaining({ name: "design", workspaceId: "ws1" }),
      );
    });
  });
});
