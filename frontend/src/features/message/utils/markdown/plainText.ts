// 抜粋用に Markdown の記号とコードブロックを取り除く
export const toPlainText = (markdown: string) =>
  markdown
    .replaceAll(/```[\s\S]*?```/g, " ")
    .replaceAll(/[*_~`>#]|^\s*[-+]\s|^\s*\d+\.\s/gm, "")
    .replaceAll(/\s+/g, " ")
    .trim();
