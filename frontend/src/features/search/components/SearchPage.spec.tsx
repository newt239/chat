import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageService } from "#/gen/chat/v1/direct_message_service_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";
import {
  ChannelSearchResultSchema,
  MessageSearchResultSchema,
  SearchHas,
  SearchService,
  SearchSort,
  SearchTarget,
  UserSearchResultSchema,
} from "#/gen/chat/v1/search_service_pb";
import { WorkspaceMemberSchema, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { SearchPage } from "./SearchPage";

import type { SearchWorkspaceRequest } from "#/gen/chat/v1/search_service_pb";

const setup = async (url: string) => {
  const search = vi.fn<(req: SearchWorkspaceRequest) => void>();
  const { router } = await renderWithProviders(<SearchPage />, url, (routes) => {
    routes.rpc(ChannelService.method.listChannels, () => ({
      channels: [
        create(ChannelSchema, { id: "c0", name: "dev" }),
        create(ChannelSchema, { id: "c1", name: "dev/frontend" }),
      ],
    }));
    routes.rpc(WorkspaceService.method.listMembers, () => ({
      members: [
        create(WorkspaceMemberSchema, {
          displayName: "Bob",
          email: "bob@example.com",
          userId: "u-bob",
        }),
        create(WorkspaceMemberSchema, {
          displayName: "Carol",
          email: "carol@example.com",
          userId: "u-carol",
        }),
      ],
    }));
    routes.rpc(DirectMessageService.method.listDirectMessages, () => ({ directMessages: [] }));
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
  test("キーワードも条件もなければ検索せずに案内を出す", async () => {
    const { search } = await setup("/app/ws1/search");
    expect(
      await screen.findByText("キーワードか条件を入力して検索してください"),
    ).toBeInTheDocument();
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

  test("修飾子を名前から ID に解決し、構造化した条件で検索する", async () => {
    const { search } = await setup(
      `/app/ws1/search?q=${encodeURIComponent("release from:@bob in:#dev has:image after:2026-09-01 before:2026-09-30")}`,
    );
    await waitFor(() => {
      expect(search).toHaveBeenCalled();
    });
    const req = search.mock.lastCall?.[0];
    expect(req?.query).toBe("release");
    expect(req?.messageFilter).toMatchObject({
      channelIds: ["c0"],
      excludeReplies: false,
      fromUserIds: ["u-bob"],
      has: [SearchHas.IMAGE],
      includeDescendantChannels: true,
    });
    const { after, before } = req?.messageFilter ?? {};
    expect(after && timestampDate(after)).toEqual(new Date(2026, 8, 1));
    expect(before && timestampDate(before)).toEqual(new Date(2026, 9, 1));

    expect(screen.getByRole("button", { name: /投稿者: Bob/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /チャンネル: #dev/ })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /期間: 2026-09-01〜2026-09-30/ }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /添付: 画像/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "下階層を含む" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
  });

  test("チップの操作を入力欄の修飾子に書き戻す", async () => {
    const { router } = await setup("/app/ws1/search?q=release");
    await userEvent.click(await screen.findByRole("button", { name: "ピン留め" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ q: "release is:pinned" });
    });
    expect(screen.getByRole("searchbox", { name: "検索キーワード" })).toHaveValue(
      "release is:pinned",
    );

    await userEvent.click(screen.getByRole("button", { name: "返信を含む" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ replies: false });
    });

    await userEvent.click(screen.getByRole("button", { name: /^投稿者/ }));
    await userEvent.click(await screen.findByRole("menuitemcheckbox", { name: "Carol" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({
        q: "release from:@Carol is:pinned",
      });
    });
  });

  test("条件をすべて解除するとキーワードだけを残す", async () => {
    const { router } = await setup("/app/ws1/search?q=release%20is:pinned&replies=false");
    await userEvent.click(await screen.findByRole("button", { name: "条件をすべて解除" }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ q: "release", replies: true });
    });
  });

  test("ヘルプの修飾子を押すと入力欄に挿入する", async () => {
    await setup("/app/ws1/search?q=release");
    await userEvent.click(screen.getByRole("button", { name: "has:image を挿入" }));
    expect(screen.getByRole("searchbox", { name: "検索キーワード" })).toHaveValue(
      "release has:image",
    );
  });

  test("見つからない名前があれば検索せずに知らせる", async () => {
    const { search } = await setup("/app/ws1/search?q=from:@nobody");
    expect(await screen.findByRole("alert")).toHaveTextContent("@nobody");
    expect(search).not.toHaveBeenCalled();
  });

  test("日付の形式が不正なら検索せずに知らせ、期間を指定し直すと外す", async () => {
    const { router, search } = await setup("/app/ws1/search?q=release%20after:2026/09/01");
    expect(await screen.findByRole("alert")).toHaveTextContent("after:2026/09/01");
    expect(search).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole("button", { name: "期間" }));
    fireEvent.change(await screen.findByLabelText("開始日"), { target: { value: "2026-09-05" } });
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ q: "release after:2026-09-05" });
    });
  });

  test("並び順とタブを切り替え、ページを送れる", async () => {
    const { router, search } = await setup("/app/ws1/search?q=release");
    await userEvent.click(await screen.findByRole("button", { name: /新しい順/ }));
    await userEvent.click(await screen.findByRole("option", { name: "関連度順" }));
    await waitFor(() => {
      expect(search).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort: SearchSort.RELEVANCE }),
      );
    });

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
