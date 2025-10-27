import { MantineProvider } from "@mantine/core";
import { cleanup, render, screen } from "@testing-library/react";
import { Provider, createStore } from "jotai";
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import type { PanelView } from "@/providers/store/ui";

import { currentChannelIdAtom } from "@/providers/store/workspace";

const mockUseChannels = vi.fn();
const mockUseMembers = vi.fn();
const mockUseWorkspaceSearchIndex = vi.fn();

vi.mock("@tanstack/react-router", () => {
  const createRouteStub = () => {
    const route = {
      options: {} as Record<string, unknown>,
      update: (config: Record<string, unknown>) => ({
        ...route,
        options: config,
      }),
    };
    return route;
  };

  const createFileRoute = () => (config: Record<string, unknown>) => ({
    ...createRouteStub(),
    options: config,
  });

  const createRootRoute = (config: Record<string, unknown>) => ({
    ...createRouteStub(),
    options: config,
  });

  const createRouter = (config: Record<string, unknown>) => ({ options: config });

  return {
    createFileRoute,
    createRootRoute,
    createRouter,
    Outlet: () => null,
    redirect: (args: Record<string, unknown>) => args,
  };
});

vi.mock("@/features/workspace/components/MemberPanel", () => ({
  MemberPanel: ({ workspaceId }: { workspaceId: string | null }) => (
    <div data-testid="member-panel" data-workspace-id={workspaceId}>
      メンバー
    </div>
  ),
}));

vi.mock("@/features/channel/hooks/useChannel", () => ({
  useChannels: (...args: unknown[]) => mockUseChannels(...args),
}));

vi.mock("@/features/workspace/hooks/useMembers", () => ({
  useMembers: (...args: unknown[]) => mockUseMembers(...args),
}));

vi.mock("@/features/search/hooks/useWorkspaceSearchIndex", () => ({
  useWorkspaceSearchIndex: (...args: unknown[]) => mockUseWorkspaceSearchIndex(...args),
}));

let WorkspaceRightSidebar: (typeof import("@/features/workspace/components/WorkspaceRightSidebar"))["WorkspaceRightSidebar"];

const renderSidebar = (view: PanelView, workspaceId = "workspace-1") => {
  const store = createStore();
  store.set(currentChannelIdAtom, "channel-1");

  return render(
    <Provider store={store}>
      <MantineProvider>
        <WorkspaceRightSidebar workspaceId={workspaceId} view={view} />
      </MantineProvider>
    </Provider>
  );
};

afterEach(() => {
  cleanup();
});

beforeAll(async () => {
  ({ WorkspaceRightSidebar } = await import(
    "@/features/workspace/components/WorkspaceRightSidebar"
  ));
});

beforeEach(() => {
  mockUseChannels.mockReturnValue({
    data: [
      {
        id: "channel-1",
        name: "general",
        description: "検索用のチャンネル",
        isPrivate: false,
      },
    ],
    isLoading: false,
    isError: false,
    error: null,
  });

  mockUseMembers.mockReturnValue({
    data: [
      {
        userId: "user-1",
        displayName: "検索太郎",
        email: "taro@example.com",
        role: "member",
        avatarUrl: null,
      },
    ],
    isLoading: false,
    isError: false,
    error: null,
  });

  mockUseWorkspaceSearchIndex.mockReturnValue({
    data: {
      channels: [
        {
          id: "channel-1",
          name: "general",
          description: "全般的な議論",
          isPrivate: false,
        },
      ],
      members: [
        {
          userId: "user-1",
          displayName: "山田太郎",
          email: "taro@example.com",
          role: "member",
          avatarUrl: null,
        },
      ],
      messages: [
        {
          id: "message-1",
          channelId: "channel-1",
          userId: "user-1",
          parentId: null,
          body: "検索対象のメッセージ",
          createdAt: new Date().toISOString(),
          editedAt: null,
          deletedAt: null,
          user: {
            id: "user-1",
            displayName: "山田太郎",
            avatarUrl: null,
            email: "taro@example.com",
          },
        },
      ],
    },
    isLoading: false,
    isError: false,
    error: null,
  });
});

describe("WorkspaceRightSidebar", () => {
  it("メンバー一覧を表示する", () => {
    renderSidebar({ type: "members" });

    expect(screen.getByTestId("member-panel")).toBeDefined();
    expect(screen.getByTestId("member-panel").getAttribute("data-workspace-id")).toBe(
      "workspace-1"
    );
  });

  it("チャンネル情報を表示する", () => {
    renderSidebar({ type: "channel-info", channelId: "channel-1" });

    expect(screen.getByText("チャンネル情報")).toBeDefined();
    expect(screen.getByText("#general")).toBeDefined();
  });

  it("スレッド情報を表示する", () => {
    renderSidebar({ type: "thread", threadId: "thread-1" });

    expect(screen.getByText("スレッド")).toBeDefined();
    expect(screen.getByText(/ID: thread-1/)).toBeDefined();
  });

  it("ユーザープロフィールを表示する", () => {
    renderSidebar({ type: "user-profile", userId: "user-1" });

    expect(screen.getByText("検索太郎")).toBeDefined();
    expect(screen.getByText("taro@example.com")).toBeDefined();
  });

  it("検索結果を表示する", () => {
    const { container } = renderSidebar({ type: "search", query: "検索", filter: "all" });

    expect(screen.getByText("検索結果")).toBeDefined();
    expect(screen.getByText((text) => text.includes("検索対象のメッセージ"))).toBeDefined();
    const cards = container.querySelectorAll(".mantine-Card-root");
    expect(cards.length).toBeGreaterThan(0);
  });

  it("非表示ビューの場合は何も描画しない", () => {
    const { container } = renderSidebar({ type: "hidden" });

    expect(container.querySelector(".border-l")).toBeNull();
    expect(container.querySelector("[data-testid='member-panel']")).toBeNull();
  });
});
