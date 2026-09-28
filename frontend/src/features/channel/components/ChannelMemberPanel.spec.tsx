import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { ChannelMemberSchema, ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import { rightSidePanelViewAtom } from "#/providers/store/ui";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelMemberPanel } from "./ChannelMemberPanel";

describe("ChannelMemberPanel", () => {
  test("メンバーを一覧し、押すとプロフィールを右パネルに開く", async () => {
    const { store } = await renderWithProviders(
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
    expect(store.get(rightSidePanelViewAtom)).toEqual({ type: "user-profile", userId: "u-bob" });
  });

  test("メンバーがいなければその旨を表示する", async () => {
    await renderWithProviders(<ChannelMemberPanel channelId="c1" />, "/app/ws1", (routes) => {
      routes.rpc(ChannelMemberService.method.listChannelMembers, () => ({ members: [] }));
    });
    expect(await screen.findByText("まだメンバーがいません")).toBeInTheDocument();
  });
});
