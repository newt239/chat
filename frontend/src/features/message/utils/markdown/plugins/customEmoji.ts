import { visit } from "unist-util-visit";

import { CUSTOM_EMOJI_IN_TEXT } from "#/features/customEmoji/utils/customEmoji";

import type { Root, RootContent, Text } from "mdast";

// :name: を span.custom-emoji にする。画像への解決は表示時に行う
export const remarkCustomEmoji = () => (tree: Root) => {
  visit(tree, "text", (node: Text, index, parent) => {
    if (!parent || index === undefined) {
      return;
    }

    const { value } = node;
    const matches = [...value.matchAll(CUSTOM_EMOJI_IN_TEXT)];
    if (matches.length === 0) {
      return;
    }

    const newNodes: RootContent[] = [];
    let lastIndex = 0;
    for (const match of matches) {
      const name = match.groups?.name;
      if (name === undefined) {
        continue;
      }
      if (match.index > lastIndex) {
        newNodes.push({ type: "text", value: value.slice(lastIndex, match.index) });
      }
      newNodes.push({
        data: {
          hName: "span",
          hProperties: { className: ["custom-emoji"], dataEmoji: name },
        },
        type: "customEmoji",
        value: name,
      });
      lastIndex = match.index + match[0].length;
    }
    if (lastIndex < value.length) {
      newNodes.push({ type: "text", value: value.slice(lastIndex) });
    }

    parent.children.splice(index, 1, ...newNodes);
  });
};
