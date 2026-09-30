import type { Messages } from "../../messages";

export const reaction: Messages["reaction"] = {
  add: "Add reaction",
  collapse: "Show less",
  list: {
    all: "All",
    open: "View reactions",
    title: "Reactions",
    undo: "Remove",
  },
  more: "+{{count}}",
  moreLabel: "Show more reactions",
  picker: {
    custom: "Custom",
  },
  names: {
    others: "{{names}} and {{count}} others",
    you: "You",
  },
  summary: "{{emoji}} {{count}}. {{names}}",
  tooltip: {
    hint: "Right-click to see who reacted and when",
    reacted: "Reacted by {{names}}",
  },
};
