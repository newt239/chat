import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

import { SearchResultList } from "./SearchResultList";

const mockNavigate = vi.fn();

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => mockNavigate,
}));

const mockChannels = [
  {
    id: "channel-1",
    workspaceId: "workspace-1",
    name: "general",
    description: "一般的な話題のチャンネル",
    isPrivate: false,
    createdBy: "user-1",
    createdAt: "2025-01-01T00:00:00Z",
  },
  {
    id: "channel-2",
    workspaceId: "workspace-1",
    name: "private-channel",
    description: "プライベートチャンネル",
    isPrivate: true,
    createdBy: "user-1",
    createdAt: "2025-01-01T00:00:00Z",
  },
];

const mockUsers = [
  {
    userId: "user-1",
    email: "test@example.com",
    displayName: "テストユーザー",
    avatarUrl: null,
    role: "owner" as const,
    joinedAt: "2025-01-01T00:00:00Z",
  },
  {
    userId: "user-2",
    email: "admin@example.com",
    displayName: "管理者",
    avatarUrl: "https://example.com/avatar.jpg",
    role: "admin" as const,
    joinedAt: "2025-01-02T00:00:00Z",
  },
];

const mockMessages = [
  {
    id: "message-1",
    channelId: "channel-1",
    userId: "user-1",
    parentId: null,
    body: "こんにちは、テストメッセージです",
    createdAt: "2025-01-01T12:00:00Z",
    editedAt: null,
    deletedAt: null,
  },
  {
    id: "message-2",
    channelId: "channel-1",
    userId: "user-2",
    parentId: null,
    body: "編集されたメッセージ",
    createdAt: "2025-01-01T13:00:00Z",
    editedAt: "2025-01-01T14:00:00Z",
    deletedAt: null,
  },
];


describe("SearchResultList", () => {
  it("チャンネルを表示する", () => {
    render(
      <SearchResultList
        messages={[]}
        channels={mockChannels}
        users={[]}
        filter="channels"
        workspaceId="workspace-1"
      />
    );

    expect(screen.getByText("チャンネル")).toBeInTheDocument();
    expect(screen.getByText("general")).toBeInTheDocument();
    expect(screen.getByText("一般的な話題のチャンネル")).toBeInTheDocument();
    expect(screen.getByText("private-channel")).toBeInTheDocument();
    expect(screen.getByText("プライベート")).toBeInTheDocument();
  });

  it("ユーザーを表示する", () => {
    render(
      <SearchResultList
        messages={[]}
        channels={[]}
        users={mockUsers}
        filter="users"
        workspaceId="workspace-1"
      />
    );

    expect(screen.getByText("ユーザー")).toBeInTheDocument();
    expect(screen.getByText("テストユーザー")).toBeInTheDocument();
    expect(screen.getByText("test@example.com")).toBeInTheDocument();
    expect(screen.getByText("管理者")).toBeInTheDocument();
    expect(screen.getByText("admin@example.com")).toBeInTheDocument();
  });

  it("メッセージを表示する", () => {
    render(
      <SearchResultList
        messages={mockMessages}
        channels={[]}
        users={[]}
        filter="messages"
        workspaceId="workspace-1"
      />
    );

    expect(screen.getByText("メッセージ")).toBeInTheDocument();
    expect(screen.getByText("こんにちは、テストメッセージです")).toBeInTheDocument();
    expect(screen.getByText("編集されたメッセージ")).toBeInTheDocument();
    expect(screen.getByText("(編集済み)")).toBeInTheDocument();
  });

  it("allフィルターで全種類を表示する", () => {
    render(
      <SearchResultList
        messages={mockMessages}
        channels={mockChannels}
        users={mockUsers}
        filter="all"
        workspaceId="workspace-1"
      />
    );

    expect(screen.getByText("チャンネル")).toBeInTheDocument();
    expect(screen.getByText("ユーザー")).toBeInTheDocument();
    expect(screen.getByText("メッセージ")).toBeInTheDocument();
  });

  it("チャンネルをクリックすると遷移する", async () => {
    const user = userEvent.setup();

    render(
      <SearchResultList
        messages={[]}
        channels={mockChannels}
        users={[]}
        filter="channels"
        workspaceId="workspace-1"
      />
    );

    const channelCard = screen.getByText("general").closest("div[class*='cursor-pointer']");
    expect(channelCard).toBeInTheDocument();

    if (channelCard) {
      await user.click(channelCard);
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/app/$workspaceId/$channelId",
        params: { workspaceId: "workspace-1", channelId: "channel-1" },
      });
    }
  });

  it("メッセージをクリックすると遷移する", async () => {
    const user = userEvent.setup();

    render(
      <SearchResultList
        messages={mockMessages}
        channels={[]}
        users={[]}
        filter="messages"
        workspaceId="workspace-1"
      />
    );

    const messageCard = screen
      .getByText("こんにちは、テストメッセージです")
      .closest("div[class*='cursor-pointer']");
    expect(messageCard).toBeInTheDocument();

    if (messageCard) {
      await user.click(messageCard);
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/app/$workspaceId/$channelId",
        params: { workspaceId: "workspace-1", channelId: "channel-1" },
        search: { message: "message-1" },
      });
    }
  });
});
