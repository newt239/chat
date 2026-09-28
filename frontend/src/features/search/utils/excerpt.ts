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
