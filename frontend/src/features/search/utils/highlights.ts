type Range = { start: number; end: number };

// 最初の一致が後ろにあるときは、その少し前から切り出して一致箇所が見えるようにする
export const excerpt = (text: string, ranges: readonly Range[], context: number) => {
  const first = ranges[0]?.start ?? 0;
  if (first <= context) {
    return { isTrimmed: false, ranges, text };
  }
  // サロゲートペアの途中で切らない
  const offset = first - context + (/[\uDC00-\uDFFF]/.test(text.charAt(first - context)) ? 1 : 0);
  return {
    isTrimmed: true,
    ranges: ranges.map(({ start, end }) => ({ end: end - offset, start: start - offset })),
    text: text.slice(offset),
  };
};

// 一致範囲（UTF-16 のオフセット）で本文を分割する。範囲は昇順・重なりなしを前提に、はみ出しは切り詰める
export const splitHighlights = (text: string, ranges: readonly Range[]) => {
  const parts: { start: number; text: string; isMatch: boolean }[] = [];
  let cursor = 0;
  for (const range of ranges) {
    const start = Math.max(cursor, Math.min(range.start, text.length));
    const end = Math.max(start, Math.min(range.end, text.length));
    if (start > cursor) {
      parts.push({ isMatch: false, start: cursor, text: text.slice(cursor, start) });
    }
    if (end > start) {
      parts.push({ isMatch: true, start, text: text.slice(start, end) });
    }
    cursor = end;
  }
  if (cursor < text.length) {
    parts.push({ isMatch: false, start: cursor, text: text.slice(cursor) });
  }
  return parts;
};
