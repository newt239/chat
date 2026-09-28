import type { Messages } from "../../messages";

export const search: Messages["search"] = {
  count: "{{count}} results",
  empty: "No results match",
  emptyHint: "Try fewer keywords",
  failed: "Couldn't load search results",
  input: "Search keywords",
  next: "Next page",
  noDescription: "No description",
  page: "Page {{page}} of {{total}}",
  placeholder: "Search messages, channels, and people",
  prev: "Previous page",
  prompt: "Enter a keyword to search",
  sections: {
    all: "All",
    channels: "Channels",
    groups: "User groups",
    messages: "Messages",
    users: "People",
  },
  showInChannel: "View in channel",
  tabs: "Search in",
  title: "Search",
};
