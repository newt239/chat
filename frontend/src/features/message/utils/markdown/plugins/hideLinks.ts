import { visit } from "unist-util-visit";

import type { Paragraph, Root } from "mdast";

const isBlank = (paragraph: Paragraph) =>
  paragraph.children.every((child) => child.type === "text" && child.value.trim() === "");

// 引用カードで中身を見せるリンクは本文から外す。外して空になった段落も消す
export const remarkHideLinks = (urls: readonly string[]) => (tree: Root) => {
  if (urls.length === 0) {
    return;
  }
  const emptied = new Set<Paragraph>();
  visit(tree, "paragraph", (paragraph: Paragraph) => {
    const kept = paragraph.children.filter(
      (child) => child.type !== "link" || !urls.includes(child.url),
    );
    if (kept.length !== paragraph.children.length) {
      paragraph.children = kept;
      if (isBlank(paragraph)) {
        emptied.add(paragraph);
      }
    }
  });
  tree.children = tree.children.filter(
    (child) => child.type !== "paragraph" || !emptied.has(child),
  );
};
