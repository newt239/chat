import { create } from "@bufbuild/protobuf";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import {
  ChannelSearchResultSchema,
  MessageSearchResultSchema,
  SearchService,
  SearchTarget,
  UserSearchResultSchema,
} from "#/gen/chat/v1/search_service_pb";
import { WorkspaceMemberSchema } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { SearchPage } from "./SearchPage";

import type { SearchWorkspaceRequest } from "#/gen/chat/v1/search_service_pb";

const setup = async (url: string) => {
  const search = vi.fn<(req: SearchWorkspaceRequest) => void>();
  const { router } = await renderWithProviders(<SearchPage />, url, (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [create(ChannelSchema, { id: "c1", name: "dev/frontend" })],
    }));
    routes.rpc(SearchService.method.searchWorkspace, (req) => {
      search(req);
      return {
        channels: create(ChannelSearchResultSchema, {
          items: [create(ChannelSchema, { id: "c1", name: "dev/frontend" })],
          perPage: 20,
          total: 1,
        }),
        messages: create(MessageSearchResultSchema, {
          items: [
            {
              highlights: [{ end: 10, start: 3 }],
              message: create(MessageSchema, {
                body: "新しいrelease手順",
                channelId: "c1",
                id: "m1",
                user: { displayName: "Bob", id: "u-bob" },
              }),
            },
          ],
          perPage: 20,
          total: 45,
        }),
        users: create(UserSearchResultSchema, {
          items: [create(WorkspaceMemberSchema, { displayName: "Release Bot", userId: "u-bot" })],
          perPage: 20,
          total: 1,
        }),
      };
    });
  });
  return { router, search };
};

describe("SearchPage", () => {
  test("キーワードがなければ検索せずに案内を出す", async () => {
    const { search } = await setup("/app/ws1/search");
    expect(await screen.findByText("キーワードを入力して検索してください")).toBeInTheDocument();
    expect(search).not.toHaveBeenCalled();
  });

  test("入力して送信すると URL を更新し、種類ごとに結果を表示する", async () => {
    const { router, search } = await setup("/app/ws1/search");
    await userEvent.type(
      screen.getByRole("searchbox", { name: "検索キーワード" }),
      "release{Enter}",
    );

    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ page: 1, q: "release" });
    });
    expect(search).toHaveBeenCalledWith(
      expect.objectContaining({ query: "release", target: SearchTarget.ALL, workspaceId: "ws1" }),
    );

    const mark = await screen.findByText("release", { selector: "mark" });
    const article = mark.closest("article");
    expect(article).not.toBeNull();
    expect(within(article ?? document.body).getByText("#dev/frontend")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "チャンネルで表示" })).toHaveAttribute(
      "href",
      "/app/ws1/c1?message=m1",
    );
    expect(screen.getByRole("link", { name: /dev\/frontend/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Release Bot/ })).toBeInTheDocument();
    expect(screen.getByText("47 件")).toBeInTheDocument();
  });

  test("タブで検索対象を切り替え、ページを送れる", async () => {
    const { router, search } = await setup("/app/ws1/search?q=release");
    await userEvent.click(await screen.findByRole("tab", { name: /メッセージ/ }));
    await waitFor(() => {
      expect(search).toHaveBeenLastCalledWith(
        expect.objectContaining({ target: SearchTarget.MESSAGES }),
      );
    });
    expect(router.state.location.search).toMatchObject({ filter: "messages" });

    expect(await screen.findByText("1 / 3 ページ")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "次のページ" }));
    await waitFor(() => {
      expect(search).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }));
    });
  });
});
