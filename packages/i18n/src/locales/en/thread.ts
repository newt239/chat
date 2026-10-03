import type { Messages } from "../../messages";

export const thread: Messages["thread"] = {
  card: {
    open: "Open thread",
    replyPlaceholder: "Reply…",
    showMore_one: "Show {{count}} more reply",
    showMore_other: "Show {{count}} more replies",
    unread: "{{count}} unread",
  },
  follow: {
    follow: "Follow thread",
    followed: "Following the thread",
    unfollow: "Unfollow thread",
    unfollowed: "Unfollowed the thread",
  },
  list: {
    emptyDescription: "Threads you post or reply in appear here",
    emptyTitle: "No threads",
    failed: "Couldn't load threads",
  },
  notFound: "Thread not found",
};
