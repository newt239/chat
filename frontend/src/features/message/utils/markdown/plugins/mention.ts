import { visit } from "unist-util-visit";

import type { Root, RootContent, Text } from "mdast";

const MENTION_REGEX = /@(?<username>\w+)/g;

export const remarkMention = () => (tree: Root) => {
  visit(tree, "text", (node: Text, index, parent) => {
    if (!parent || index === undefined) {
      return;
    }

    const { value } = node;
    const matches = [...value.matchAll(MENTION_REGEX)];

    if (matches.length === 0) {
      return;
    }

    const newNodes: RootContent[] = [];
    let lastIndex = 0;

    for (const match of matches) {
      const matchIndex = match.index;
      const username = match.groups?.username;
      if (username === undefined) {
        continue;
      }

      // メンション前のテキスト
      if (matchIndex > lastIndex) {
        newNodes.push({
          type: "text",
          value: value.slice(lastIndex, matchIndex),
        });
      }

      // メンションノード
      newNodes.push({
        data: {
          hName: "span",
          hProperties: {
            className: ["mention"],
            "data-mention": username,
          },
        },
        type: "mention",
        value: username,
      });

      lastIndex = matchIndex + match[0].length;
    }

    // 残りのテキスト
    if (lastIndex < value.length) {
      newNodes.push({
        type: "text",
        value: value.slice(lastIndex),
      });
    }

    // ノードを置き換え
    parent.children.splice(index, 1, ...newNodes);
  });
};
