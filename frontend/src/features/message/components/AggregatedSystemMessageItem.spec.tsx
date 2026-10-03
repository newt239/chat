import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { SystemMessageKind, SystemMessageSchema } from "#/gen/chat/v1/message_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AggregatedSystemMessageItem } from "./AggregatedSystemMessageItem";

const joined = (channelId: string) =>
  create(SystemMessageSchema, {
    actorId: "u-bob",
    channelId,
    kind: SystemMessageKind.MEMBER_JOINED,
    payload: { userId: "u-bob" },
  });

const setup = (channelId: string) =>
  renderWithProviders(
    <AggregatedSystemMessageItem workspaceId="ws1" parentName="dev" message={joined(channelId)} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({
        channels: [
          create(ChannelSchema, { id: "dev", isMember: true, name: "dev" }),
          create(ChannelSchema, {
            id: "fe",
            isMember: true,
            name: "dev/frontend",
            parentId: "dev",
          }),
        ],
      }));
      routes.rpc(WorkspaceService.method.listMembers, () => ({
        members: [create(WorkspaceMemberSchema, { displayName: "Bob", userId: "u-bob" })],
      }));
    },
  );

describe("AggregatedSystemMessageItem", () => {
  test("下階層への参加は親からの相対パスを添える", async () => {
    await setup("fe");
    expect(await screen.findByText("#frontend に Bob が参加しました")).toBeInTheDocument();
  });

  test("親チャンネル自身への参加はそのチャンネル名を添える", async () => {
    await setup("dev");
    expect(await screen.findByText("#dev に Bob が参加しました")).toBeInTheDocument();
  });
});
