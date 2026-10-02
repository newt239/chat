import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ThreadFollowButton } from "./ThreadFollowButton";

describe("ThreadFollowButton", () => {
  test("フォロー中なら押下状態で表し、押すとフォローを解除する", async () => {
    const unfollow = vi.fn(() => ({}));
    await renderWithProviders(
      <ThreadFollowButton threadId="t1" isFollowing />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ThreadService.method.unfollowThread, unfollow);
      },
    );

    const button = await screen.findByRole("button", { name: "スレッドをフォロー" });
    expect(button).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(button);
    await waitFor(() => {
      expect(unfollow).toHaveBeenCalledWith(
        expect.objectContaining({ messageId: "t1" }),
        expect.anything(),
      );
    });
  });
});
