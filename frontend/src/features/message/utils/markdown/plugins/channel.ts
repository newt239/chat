import { visit } from "unist-util-visit";

import type { Root, RootContent, Text } from "mdast";

const CHANNEL_REGEX = /#(?<channelName>[\w-]+)/g;

export const remarkChannel = () => (tree: Root) => {
  visit(tree, "text", (node: Text, index, parent) => {
    if (!parent || index === undefined) {
      return;
    }

    const { value } = node;
    const matches = [...value.matchAll(CHANNEL_REGEX)];

    if (matches.length === 0) {
      return;
    }

    const newNodes: RootContent[] = [];
    let lastIndex = 0;

    for (const match of matches) {
      const matchIndex = match.index;
      const channelName = match.groups?.channelName;
      if (channelName === undefined) {
        continue;
      }

      // チャンネルリンク前のテキスト
      if (matchIndex > lastIndex) {
        newNodes.push({
          type: "text",
          value: value.slice(lastIndex, matchIndex),
        });
      }

      // チャンネルリンクノード
      newNodes.push({
        data: {
          hName: "span",
          hProperties: {
            className: ["channel-link"],
            "data-channel": channelName,
          },
        },
        type: "channelLink",
        value: channelName,
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
