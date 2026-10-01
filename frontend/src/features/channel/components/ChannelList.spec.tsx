import { create } from "@bufbuild/protobuf";
import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import {
  ChannelCategorySchema,
  ChannelCategoryService,
} from "#/gen/chat/v1/channel_category_service_pb";
import { ChannelSchema, ChannelService } from "#/gen/chat/v1/channel_service_pb";
import { preferencesAtom, defaultPreferences } from "#/providers/store/preferences";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelList } from "./ChannelList";

const channels = [
  create(ChannelSchema, { id: "dev", isMember: true, name: "dev" }),
  create(ChannelSchema, {
    hasMention: true,
    id: "fe",
    isMember: true,
    mentionCount: 2,
    name: "dev/frontend",
    parentId: "dev",
    unreadCount: 2,
  }),
  create(ChannelSchema, {
    id: "web",
    isMember: true,
    name: "dev/frontend/web",
    parentId: "fe",
    unreadCount: 3,
  }),
  create(ChannelSchema, {
    id: "g",
    isMember: true,
    lastMessageAt: timestampFromMs(1000),
    name: "general",
  }),
];

const workCategory = create(ChannelCategorySchema, {
  channelIds: ["fe"],
  id: "work",
  name: "仕事",
});

const render = (categoryId: string | null = null) =>
  renderWithProviders(
    <ChannelList workspaceId="ws1" categoryId={categoryId} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(ChannelService.method.listChannels, () => ({ channels }));
      routes.rpc(ChannelCategoryService.method.listChannelCategories, () => ({
        categories: categoryId === null ? [] : [workCategory],
      }));
    },
  );

describe("ChannelList", () => {
  test("階層をツリーで並べ、行には末尾の名前だけを出す", async () => {
    await render();
    expect(await screen.findByRole("link", { name: /^公開チャンネル\s?dev$/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?frontend/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?web$/ })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^公開チャンネル\s?general$/ })).toBeInTheDocument();
  });

  test("折りたたむと子の数と子孫のメンション数の合計を親に出す", async () => {
    await render();
    await userEvent.click(await screen.findByRole("button", { name: "dev の下階層を折りたたむ" }));

    const parent = await screen.findByRole("link", { name: /^公開チャンネル\s?dev/ });
    expect(parent).toHaveTextContent("1");
    expect(parent).toHaveTextContent("2");
    // 折りたたみのアニメーションが終わってから消える
    await waitFor(() => {
      expect(screen.queryByRole("link", { name: /frontend/ })).not.toBeInTheDocument();
    });

    await userEvent.click(screen.getByRole("button", { name: "dev の下階層を展開" }));
    expect(await screen.findByRole("link", { name: /frontend/ })).toBeInTheDocument();
  });

  test("自分のカテゴリに入れたチャンネルは子孫ごとそのカテゴリに出し、フルパスで表示する", async () => {
    await render("work");
    const links = await screen.findAllByRole("link");
    expect(links.map((link) => link.textContent)).toEqual(["dev / frontend2", "web"]);
  });

  test("新しいメッセージ順では階層を分けてフルパスで並べる", async () => {
    const { store } = await render();
    store.set(preferencesAtom, { ...defaultPreferences, channelSortOrder: "recentActivity" });
    const links = await screen.findAllByRole("link");
    expect(links.map((link) => link.textContent)).toEqual([
      "general",
      "dev",
      // メンションを含む未読の件数
      "dev / frontend2",
      "dev / frontend / web",
    ]);
  });
});
