import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { Mention } from "./Mention";

const setup = (value: string) =>
  renderWithProviders(<Mention data-mention={value} />, "/app/ws1", (routes) => {
    routes.rpc(WorkspaceService.method.listMembers, () => ({
      members: [create(WorkspaceMemberSchema, { displayName: "Bob Smith", userId: "u-bob" })],
    }));
    routes.rpc(UserGroupService.method.listUserGroups, () => ({
      userGroups: [create(UserGroupSchema, { id: "g1", name: "developers" })],
    }));
  });

describe("Mention", () => {
  test("ユーザーは ID で引いた今の表示名のボタンにする", async () => {
    await setup("user:u-bob");
    expect(await screen.findByRole("button", { name: "@Bob Smith" })).toBeInTheDocument();
  });

  test("グループは ID で引いた今の名前のボタンにする", async () => {
    await setup("group:g1");
    expect(await screen.findByRole("button", { name: "@developers" })).toBeInTheDocument();
  });

  test("@channel / @here はボタンにしない", async () => {
    await setup("broadcast:here");
    expect(await screen.findByText("@here")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  test("見つからないユーザーは名前を伏せる", async () => {
    await setup("user:u-nobody");
    expect(await screen.findByText("@不明なユーザー")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
