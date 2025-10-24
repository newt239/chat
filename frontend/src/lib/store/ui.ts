import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

export type RightSidebarView =
  | { type: "hidden" }
  | { type: "members" }
  | { type: "channel-info"; channelId?: string | null }
  | { type: "thread"; threadId: string }
  | { type: "user-profile"; userId: string }
  | { type: "search"; query: string; filter: "all" | "messages" | "channels" | "users" };

const defaultRightSidebarView: RightSidebarView = { type: "hidden" };

export const rightSidebarViewAtom = atomWithStorage<RightSidebarView>(
  "ui-storage:rightSidebarView",
  defaultRightSidebarView
);

const isSameRightSidebarView = (first: RightSidebarView, second: RightSidebarView): boolean => {
  if (first.type !== second.type) {
    return false;
  }

  switch (first.type) {
    case "hidden":
    case "members":
      return true;
    case "channel-info":
      return second.type === "channel-info" && first.channelId === second.channelId;
    case "thread":
      return second.type === "thread" && first.threadId === second.threadId;
    case "user-profile":
      return second.type === "user-profile" && first.userId === second.userId;
    case "search":
      return (
        second.type === "search" &&
        first.query === second.query &&
        first.filter === second.filter
      );
  }
};

export const setRightSidebarViewAtom = atom(null, (_get, set, view: RightSidebarView) => {
  set(rightSidebarViewAtom, view);
});

export const toggleRightSidebarViewAtom = atom(null, (get, set, view: RightSidebarView) => {
  const current = get(rightSidebarViewAtom);
  if (isSameRightSidebarView(current, view)) {
    set(rightSidebarViewAtom, defaultRightSidebarView);
    return;
  }
  set(rightSidebarViewAtom, view);
});

export const closeRightSidebarAtom = atom(null, (_get, set) => {
  set(rightSidebarViewAtom, defaultRightSidebarView);
});
