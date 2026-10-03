import type { Messages } from "../../messages";

export const thread: Messages["thread"] = {
  card: {
    open: "Open thread",
    replyPlaceholder: "Reply…",
    showMore: "Show {{count}} more replies",
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
  },
  notFound: "Thread not found",
};
