import { create } from "@bufbuild/protobuf";
import { screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import {
  ChannelSearchResultSchema,
  SearchService,
  SearchTarget,
} from "#/gen/chat/v1/search_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { SearchResultsPanel } from "./SearchResultsPanel";

import type { SearchWorkspaceRequest } from "#/gen/chat/v1/search_service_pb";

const render = async (query: string) => {
  const search = vi.fn<(req: SearchWorkspaceRequest) => void>();
  await renderWithProviders(
    <SearchResultsPanel workspaceId="ws1" query={query} filter="channels" />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({ channels: [] }));
      routes.rpc(SearchService.method.searchWorkspace, (req) => {
        search(req);
        return {
          channels: create(ChannelSearchResultSchema, {
            items: [create(ChannelSchema, { description: "告知用", id: "c1", name: "announce" })],
            total: 1,
          }),
        };
      });
    },
  );
  return search;
};

describe("SearchResultsPanel", () => {
  test("1 ページ目の結果を表示する", async () => {
    const search = await render("ann");
    expect(await screen.findByRole("link", { name: /announce/ })).toHaveTextContent("告知用");
    expect(search).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, query: "ann", target: SearchTarget.CHANNELS }),
    );
  });

  test("キーワードが空なら案内を出す", async () => {
    const search = await render(" ");
    expect(screen.getByText("キーワードを入力して検索してください")).toBeInTheDocument();
    expect(search).not.toHaveBeenCalled();
  });
});
