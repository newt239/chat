import type { Messages } from "../../messages";

export const dm: Messages["dm"] = {
  create: {
    callout: "More than 10 people, so this will be created as a private channel.",
    count: "{{count}} / {{max}} people including you",
    limitHint: "DMs can have up to {{max}} people. Larger groups become a private channel",
    members: "Members",
    noResults: "No matching members",
    removeSelected: "Remove {{name}}",
    search: "Search members",
    selectAtLeastOne: "Choose at least one person",
    submitChannel: "Create private channel",
    submitDM: "Start DM",
    submitGroup: "Start group DM",
    title: "Start a direct message",
  },
  groupCount: "Group DM with {{count}} people",
};
