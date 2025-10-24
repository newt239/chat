import { visit } from 'unist-util-visit';

import type { Root, Text, Parent } from 'mdast';

const MENTION_REGEX = /@(\w+)/g;

export function remarkMention() {
  return (tree: Root) => {
    visit(tree, 'text', (node: Text, index, parent) => {
      if (!parent || index === undefined) return;

      const value = node.value;
      const matches = [...value.matchAll(MENTION_REGEX)];

      if (matches.length === 0) return;

      const newNodes: unknown[] = [];
      let lastIndex = 0;

      matches.forEach((match) => {
        const matchIndex = match.index!;
        const username = match[1];

        // メンション前のテキスト
        if (matchIndex > lastIndex) {
          newNodes.push({
            type: 'text',
            value: value.slice(lastIndex, matchIndex),
          });
        }

        // メンションノード
        newNodes.push({
          type: 'mention',
          value: username,
          data: {
            hName: 'span',
            hProperties: {
              className: ['mention'],
              'data-mention': username,
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
