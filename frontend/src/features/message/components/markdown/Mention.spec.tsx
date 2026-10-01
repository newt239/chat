import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { UserGroupSchema, UserGroupService } from "#/gen/chat/v1/user_group_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { Mention } from "./Mention";

const setup = (username: string) =>
  renderWithProviders(<Mention data-mention={username} />, "/app/ws1", (routes) => {
    routes.rpc(WorkspaceService.method.listMembers, () => ({
      members: [create(WorkspaceMemberSchema, { displayName: "Bob Smith", userId: "u-bob" })],
    }));
    routes.rpc(UserGroupService.method.listUserGroups, () => ({
      userGroups: [create(UserGroupSchema, { id: "g1", name: "developers" })],
    }));
  });

describe("Mention", () => {
  test("表示名の前方一致で見つかったユーザーは大文字小文字を問わずボタンにする", async () => {
    await setup("bob");
    expect(await screen.findByRole("button", { name: "@bob" })).toBeInTheDocument();
  });

  test("名前が一致するグループはボタンにする", async () => {
    await setup("developers");
    expect(await screen.findByRole("button", { name: "@developers" })).toBeInTheDocument();
  });

  test("存在しない名前はハイライトせず文字列のまま出す", async () => {
    await setup("nobody");
    expect(await screen.findByText("@nobody")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
