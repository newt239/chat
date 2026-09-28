import { createStore } from "jotai";
import { beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { registerRouter } from "#/lib/navigation";

import { openThreadRouteAtom, rightSidePanelViewAtom, setRightSidePanelViewAtom } from "./ui";
import { setCurrentChannelAtom, syncCurrentWorkspaceAtom } from "./workspace";

const navigate = vi.fn(async () => {});

describe("setRightSidePanelViewAtom", () => {
  beforeEach(() => {
    navigate.mockClear();
    registerRouter({ navigate });
  });

  test("スレッドは表示中のチャンネルのスレッド URL へ遷移する", () => {
    const store = createStore();
    store.set(syncCurrentWorkspaceAtom, "ws");
    store.set(setCurrentChannelAtom, "ch");

    store.set(setRightSidePanelViewAtom, { threadId: "m1", type: "thread" });

    expect(navigate).toHaveBeenCalledWith({
      params: { channelId: "ch", messageId: "m1", workspaceId: "ws" },
      to: "/app/$workspaceId/$channelId/thread/$messageId",
    });
    expect(store.get(rightSidePanelViewAtom)).toEqual({ type: "hidden" });
  });

  test("スレッドを開いているときに他のパネルを開くとスレッドを閉じる", () => {
    const store = createStore();
    store.set(openThreadRouteAtom, { channelId: "ch", messageId: "m1", workspaceId: "ws" });

    store.set(setRightSidePanelViewAtom, { type: "user-profile", userId: "u1" });

    expect(store.get(rightSidePanelViewAtom)).toEqual({ type: "user-profile", userId: "u1" });
    expect(navigate).toHaveBeenCalledWith({
      params: { channelId: "ch", workspaceId: "ws" },
      to: "/app/$workspaceId/$channelId",
    });
  });

  test("閉じるだけならスレッドの URL はそのまま", () => {
    const store = createStore();
    store.set(openThreadRouteAtom, { channelId: "ch", messageId: "m1", workspaceId: "ws" });

    store.set(setRightSidePanelViewAtom, { type: "hidden" });

    expect(navigate).not.toHaveBeenCalled();
  });
});
