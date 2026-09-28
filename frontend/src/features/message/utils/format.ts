type FormatPattern = {
  prefix: string;
  suffix: string;
};

export type Selection = { start: number; end: number };

const formatPatterns = {
  bold: { prefix: "**", suffix: "**" },
  code: { prefix: "`", suffix: "`" },
  heading: { prefix: "# ", suffix: "" },
  italic: { prefix: "_", suffix: "_" },
  link: { prefix: "[", suffix: "](url)" },
  list: { prefix: "- ", suffix: "" },
  orderedList: { prefix: "1. ", suffix: "" },
  quote: { prefix: "> ", suffix: "" },
  strikethrough: { prefix: "~~", suffix: "~~" },
} satisfies Record<string, FormatPattern>;

export type FormatKey = keyof typeof formatPatterns;

// 選択範囲を書式の記号で囲み、カーソルを選択範囲の末尾（記号の内側）に置く
export const applyFormat = (text: string, { start, end }: Selection, key: FormatKey) => {
  const { prefix, suffix } = formatPatterns[key];
  return {
    cursor: end + prefix.length,
    text: text.slice(0, start) + prefix + text.slice(start, end) + suffix + text.slice(end),
  };
};

// カーソル位置に効いている書式
export const detectActiveFormats = (text: string, { start, end }: Selection) => {
  const before = text.slice(0, start);
  const after = text.slice(end);
  const lineStart = before.lastIndexOf("\n") + 1;
  const lineEnd = text.indexOf("\n", end);
  const line = text.slice(lineStart, lineEnd === -1 ? text.length : lineEnd).trimStart();

  return {
    bold: /\*\*[^*]*$/.test(before) && /^[^*]*\*\*/.test(after),
    code: /`[^`]*$/.test(before) && /^[^`]*`/.test(after),
    heading: line.startsWith("#"),
    italic: /_[^_]*$/.test(before) && /^[^_]*_/.test(after),
    link: /\[[^\]]*$/.test(before) && /^[^\]]*\]/.test(after),
    list: /^-\s/.test(line),
    orderedList: /^\d+\.\s/.test(line),
    quote: line.startsWith(">"),
    strikethrough: /~~[^~]*$/.test(before) && /^[^~]*~~/.test(after),
  } satisfies Record<FormatKey, boolean>;
};
