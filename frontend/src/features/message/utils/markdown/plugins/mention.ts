import { visit } from "unist-util-visit";

import { splitMentionTokens, toMentionToken } from "#/features/mention/utils/mentionToken";

import type { MentionPart } from "#/features/mention/utils/mentionToken";

import type { Root, RootContent, Text } from "mdast";

const toNode = (part: MentionPart) => {
  if (part.kind === "text") {
    return { type: "text", value: part.text } satisfies RootContent;
  }
  // sanitize は hast のプロパティ名（キャメルケース）で判定する
  return {
    data: {
      hName: "span",
      hProperties: { className: ["mention"], dataMention: toMentionToken(part) },
    },
    type: "mention",
    value: part.id,
  } satisfies RootContent;
};

/** 本文に ID で埋め込んだメンションとチャンネルを、今の名前で描画するノードにする */
export const remarkMention = () => (tree: Root) => {
  visit(tree, "text", (node: Text, index, parent) => {
    if (!parent || index === undefined) {
      return;
    }
    const parts = splitMentionTokens(node.value);
    if (parts.some((part) => part.kind !== "text")) {
      parent.children.splice(index, 1, ...parts.map((part) => toNode(part)));
    }
  });
};
