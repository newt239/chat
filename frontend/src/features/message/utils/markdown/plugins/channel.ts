import { visit } from 'unist-util-visit';

import type { Root, Text, Parent } from 'mdast';

const CHANNEL_REGEX = /#([\w-]+)/g;

export function remarkChannel() {
  return (tree: Root) => {
    visit(tree, 'text', (node: Text, index, parent) => {
      if (!parent || index === undefined) return;

      const value = node.value;
      const matches = [...value.matchAll(CHANNEL_REGEX)];

      if (matches.length === 0) return;

      const newNodes: unknown[] = [];
      let lastIndex = 0;

      matches.forEach((match) => {
        const matchIndex = match.index!;
        const channelName = match[1];

        // チャンネルリンク前のテキスト
        if (matchIndex > lastIndex) {
          newNodes.push({
            type: 'text',
            value: value.slice(lastIndex, matchIndex),
          });
        }

        // チャンネルリンクノード
        newNodes.push({
          type: 'channelLink',
          value: channelName,
          data: {
            hName: 'span',
            hProperties: {
              className: ['channel-link'],
              'data-channel': channelName,
            },
          },
        });

        lastIndex = matchIndex + match[0].length;
      });

      // 残りのテキスト
      if (lastIndex < value.length) {
        newNodes.push({
          type: 'text',
          value: value.slice(lastIndex),
        });
      }

      // ノードを置き換え
      const parentNode = parent as Parent;
      parentNode.children.splice(index, 1, ...newNodes as Text[]);
    });
  };
}
