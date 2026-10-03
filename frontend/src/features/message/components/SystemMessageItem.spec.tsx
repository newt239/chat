import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { SystemMessageKind, SystemMessageSchema } from "#/gen/chat/v1/message_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { SystemMessageItem } from "./SystemMessageItem";

const pinned = (payload: Record<string, string>) =>
  create(SystemMessageSchema, {
    actorId: "u-bob",
    channelId: "ch1",
    kind: SystemMessageKind.MESSAGE_PINNED,
    payload,
  });

const setup = (payload: Record<string, string>) =>
  renderWithProviders(<SystemMessageItem message={pinned(payload)} />, "/app/ws1", (routes) => {
    routes.rpc(WorkspaceService.method.listMembers, () => ({
      members: [create(WorkspaceMemberSchema, { displayName: "Bob", userId: "u-bob" })],
    }));
  });

describe("SystemMessageItem", () => {
  test("ピン留めのお知らせはピン留めされたメッセージへリンクする", async () => {
    await setup({ messageId: "m1" });
    expect(await screen.findByRole("link", { name: "メッセージ" })).toHaveAttribute(
      "href",
      "/app/ws1/ch1?message=m1",
    );
  });

  test("スレッドの返信はスレッドを開くリンクにする", async () => {
    await setup({ messageId: "r1", parentId: "m1" });
    expect(await screen.findByRole("link", { name: "メッセージ" })).toHaveAttribute(
      "href",
      "/app/ws1/ch1/thread/m1?message=r1",
    );
  });
});
