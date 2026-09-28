import { create } from "@bufbuild/protobuf";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelMemberSchema, ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { channelViewersAtom } from "#/providers/store/ui";
import { syncCurrentWorkspaceAtom } from "#/providers/store/workspace";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelMemberPanel } from "./ChannelMemberPanel";

describe("ChannelMemberPanel", () => {
  test("メンバーを一覧し、押すとプロフィールを右パネルに開く", async () => {
    const { router } = await renderWithProviders(
      <ChannelMemberPanel channelId="c1" />,
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
      <ChannelMemberPanel channelId="c1" />,
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
    store.set(syncCurrentWorkspaceAtom, "ws1");
    store.set(channelViewersAtom, { c1: ["u-bob"] });

    const viewing = await screen.findByRole("region", { name: "いま閲覧中" });
    expect(await within(viewing).findByText("ボブさん")).toBeInTheDocument();
    expect(within(viewing).getByText("閲覧中")).toBeInTheDocument();
    const others = screen.getByRole("region", { name: "その他のメンバー" });
    expect(within(others).getByText("Carol")).toBeInTheDocument();
  });

  test("メンバーがいなければその旨を表示する", async () => {
    await renderWithProviders(<ChannelMemberPanel channelId="c1" />, "/app/ws1", (routes) => {
      routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({ members: [] }));
    });
    expect(await screen.findByText("まだメンバーがいません")).toBeInTheDocument();
  });
});
