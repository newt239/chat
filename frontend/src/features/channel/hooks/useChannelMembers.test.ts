import { renderHook, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import {
  useChannelMembers,
  useInviteChannelMember,
  useJoinChannel,
  useLeaveChannel,
  useRemoveChannelMember,
  useUpdateChannelMemberRole,
} from "./useChannelMembers";

import { api } from "@/lib/api/client";
import { createAppWrapper, createTestQueryClient } from "@/test/utils";

const createWrapper = () => createAppWrapper(createTestQueryClient());

vi.mock("@/lib/api/client", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
    PATCH: vi.fn(),
    DELETE: vi.fn(),
  },
}));

describe("useChannelMembers", () => {
  it("fetches channel members successfully", async () => {
    const mockMembers = [
      {
        userId: "user-1",
        email: "user1@example.com",
        displayName: "User 1",
        avatarUrl: null,
        role: "admin",
        joinedAt: "2024-01-01T00:00:00Z",
      },
    ];

    vi.mocked(api.GET).mockResolvedValueOnce({
      data: { members: mockMembers },
      error: undefined,
    });

    const { result } = renderHook(() => useChannelMembers("channel-1"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockMembers);
  });

  it("returns empty array when channelId is null", () => {
    const { result } = renderHook(() => useChannelMembers(null), {
      wrapper: createWrapper(),
    });

    expect(result.current.data).toBeUndefined();
    expect(result.current.isLoading).toBe(false);
  });
});

describe("useInviteChannelMember", () => {
  it("invites member successfully", async () => {
    vi.mocked(api.POST).mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: {} as Response,
    });

    const { result } = renderHook(() => useInviteChannelMember("channel-1"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ userId: "user-1", role: "member" });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});

describe("useJoinChannel", () => {
  it("joins channel successfully", async () => {
    vi.mocked(api.POST).mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: {} as Response,
    });

    const { result } = renderHook(() => useJoinChannel("channel-1"), {
      wrapper: createWrapper(),
    });

    result.current.mutate();

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});

describe("useLeaveChannel", () => {
  it("leaves channel successfully", async () => {
    vi.mocked(api.DELETE).mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: {} as Response,
    });

    const { result } = renderHook(() => useLeaveChannel("channel-1"), {
      wrapper: createWrapper(),
    });

    result.current.mutate();

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});

describe("useUpdateChannelMemberRole", () => {
  it("updates member role successfully", async () => {
    vi.mocked(api.PATCH).mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: {} as Response,
    });

    const { result } = renderHook(() => useUpdateChannelMemberRole("channel-1"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ userId: "user-1", role: "admin" });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});

describe("useRemoveChannelMember", () => {
  it("removes member successfully", async () => {
    vi.mocked(api.DELETE).mockResolvedValueOnce({
      data: undefined,
      error: undefined,
      response: {} as Response,
    });

    const { result } = renderHook(() => useRemoveChannelMember("channel-1"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ userId: "user-1" });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});
