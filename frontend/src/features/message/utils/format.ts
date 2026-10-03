export type Selection = { start: number; end: number };

// 行内の書式。付けるときは prefix と suffix で囲み、外すときは tokens に一致する記号の対を探す
const inlineFormats = {
  bold: { prefix: "**", suffix: "**", tokens: [/(?<!\*)\*\*(?!\*)/g, /(?<!_)__(?!_)/g] },
  code: { prefix: "`", suffix: "`", tokens: [/`/g] },
  italic: {
    prefix: "_",
    suffix: "_",
    tokens: [/(?<![*\w])\*(?!\*)|(?<!\*)\*(?![*\w])/g, /(?<![_\w])_(?!_)|(?<!_)_(?![_\w])/g],
  },
  strikethrough: { prefix: "~~", suffix: "~~", tokens: [/~~/g] },
};

const headingMarker = /^#{1,6} /;
const quoteMarker = /^> ?/;
// 箇条書きと番号付きは付けるときに互いを置き換える
const listMarker = /^(?:[-*+] (?:\[[ xX]\] )?|\d+\. )/;

// 行頭の書式。marker が付いていれば外し、なければ replaces を取り除いて prefix を付ける
const lineFormats = {
  heading: { marker: headingMarker, prefix: () => "# ", replaces: headingMarker },
  list: { marker: /^[-*+] (?:\[[ xX]\] )?/, prefix: () => "- ", replaces: listMarker },
  orderedList: {
    marker: /^\d+\. /,
    prefix: (index: number) => `${index + 1}. `,
    replaces: listMarker,
  },
  quote: { marker: quoteMarker, prefix: () => "> ", replaces: quoteMarker },
};

const lineRangeOf = (text: string, { start, end }: Selection) => {
  const lineEnd = text.indexOf("\n", end);
  return {
    end: lineEnd === -1 ? text.length : lineEnd,
    start: text.lastIndexOf("\n", start - 1) + 1,
  };
};

// 選択範囲を内側に含む、または選択範囲とちょうど重なる記号の対
const findEnclosingPair = (text: string, selection: Selection, tokens: RegExp[]) => {
  const line = lineRangeOf(text, selection);
  const lineText = text.slice(line.start, line.end);
  const pairs = tokens.flatMap((token) => {
    const found = [...lineText.matchAll(token)].map((match) => ({
      end: line.start + match.index + match[0].length,
      start: line.start + match.index,
    }));
    // 記号は左から順に開きと閉じの対になる
    return found.flatMap((open, index) => {
      const close = found[index + 1];
      return index % 2 === 0 && close ? [{ close, open }] : [];
    });
  });
  return (
    pairs.find(
      ({ open, close }) =>
        (open.end <= selection.start && close.start >= selection.end) ||
        (open.start === selection.start && close.end === selection.end),
    ) ?? null
  );
};

const findEnclosingLink = (text: string, selection: Selection) => {
  const line = lineRangeOf(text, selection);
  for (const match of text.slice(line.start, line.end).matchAll(/\[(?<label>[^\]]*)\]\([^)]*\)/g)) {
    const start = line.start + match.index;
    const end = start + match[0].length;
    if (start <= selection.start && end >= selection.end) {
      return { end, label: match.groups?.label ?? "", start };
    }
  }
  return null;
};

const removePair = (
  text: string,
  selection: Selection,
  { open, close }: { open: Selection; close: Selection },
) => {
  const openLength = open.end - open.start;
  const shift = (position: number) =>
    Math.min(Math.max(position, open.end), close.start) - openLength;
  return {
    selection: { end: shift(selection.end), start: shift(selection.start) },
    text: text.slice(0, open.start) + text.slice(open.end, close.start) + text.slice(close.end),
  };
};

const wrap = (
  text: string,
  { start, end }: Selection,
  { prefix, suffix }: { prefix: string; suffix: string },
) => ({
  selection: { end: end + prefix.length, start: start + prefix.length },
  text: text.slice(0, start) + prefix + text.slice(start, end) + suffix + text.slice(end),
});

const toggleLink = (text: string, selection: Selection) => {
  const link = findEnclosingLink(text, selection);
  if (link) {
    return {
      selection: { end: link.start + link.label.length, start: link.start },
      text: text.slice(0, link.start) + link.label + text.slice(link.end),
    };
  }
  const inserted = wrap(text, selection, { prefix: "[", suffix: "](url)" });
  // 続けて URL を貼れるよう url を選択する
  const urlStart = inserted.selection.end + 2;
  return { selection: { end: urlStart + 3, start: urlStart }, text: inserted.text };
};

// 選択範囲にかかる各行の行頭へ記号を付ける。すべての行に付いていれば外す
const toggleLine = (key: keyof typeof lineFormats) => (text: string, selection: Selection) => {
  const { marker, prefix, replaces } = lineFormats[key];
  const range = lineRangeOf(text, selection);
  const lines = text.slice(range.start, range.end).split("\n");
  const targets = lines.filter((line) => line.trim() !== "" || lines.length === 1);
  const isActive = targets.every((line) => marker.test(line));
  let count = 0;
  const nextLines = lines.map((line) => {
    if (!targets.includes(line)) {
      return line;
    }
    if (isActive) {
      return line.replace(marker, "");
    }
    const next = prefix(count) + line.replace(replaces, "");
    count += 1;
    return next;
  });
  const block = nextLines.join("\n");
  const nextText = text.slice(0, range.start) + block + text.slice(range.end);
  if (lines.length > 1 || selection.start !== selection.end) {
    return { selection: { end: range.start + block.length, start: range.start }, text: nextText };
  }
  const cursor = Math.max(range.start, selection.start + block.length - (range.end - range.start));
  return { selection: { end: cursor, start: cursor }, text: nextText };
};

const toggleInline = (key: keyof typeof inlineFormats) => (text: string, selection: Selection) => {
  const pattern = inlineFormats[key];
  const pair = findEnclosingPair(text, selection, pattern.tokens);
  return pair ? removePair(text, selection, pair) : wrap(text, selection, pattern);
};

const formatToggles = {
  bold: toggleInline("bold"),
  code: toggleInline("code"),
  heading: toggleLine("heading"),
  italic: toggleInline("italic"),
  link: toggleLink,
  list: toggleLine("list"),
  orderedList: toggleLine("orderedList"),
  quote: toggleLine("quote"),
  strikethrough: toggleInline("strikethrough"),
};

export type FormatKey = keyof typeof formatToggles;

// 書式を切り替える。既に効いていれば記号を外し、なければ付ける
export const toggleFormat = (text: string, selection: Selection, key: FormatKey) =>
  formatToggles[key](text, selection);

// 選択範囲を絵文字で置き換える。:name: は英数字と接すると絵文字にならないため空白を挟む
export const insertEmoji = (text: string, { start, end }: Selection, emoji: string) => {
  const isShortcode = emoji.startsWith(":");
  const before = text.slice(0, start);
  const after = text.slice(end);
  const leading = isShortcode && /\w$/u.test(before) ? " " : "";
  const trailing = isShortcode && /^\w/u.test(after) ? " " : "";
  const inserted = leading + emoji + trailing;
  return { cursor: start + inserted.length, text: before + inserted + after };
};

// カーソル位置に効いている書式
export const detectActiveFormats = (text: string, selection: Selection) => {
  const lineRange = lineRangeOf(text, { end: selection.start, start: selection.start });
  const line = text.slice(lineRange.start, lineRange.end);
  const isInline = (key: keyof typeof inlineFormats) =>
    findEnclosingPair(text, selection, inlineFormats[key].tokens) !== null;
  return {
    bold: isInline("bold"),
    code: isInline("code"),
    heading: lineFormats.heading.marker.test(line),
    italic: isInline("italic"),
    link: findEnclosingLink(text, selection) !== null,
    list: lineFormats.list.marker.test(line),
    orderedList: lineFormats.orderedList.marker.test(line),
    quote: lineFormats.quote.marker.test(line),
    strikethrough: isInline("strikethrough"),
  } satisfies Record<FormatKey, boolean>;
};

// 箇条書き・番号付き・タスク・引用の行で改行したとき、次の行に記号を引き継ぐ。項目が空なら記号を消してリストを抜ける
export const continueList = (text: string, { start, end }: Selection) => {
  if (start !== end) {
    return null;
  }
  const lineStart = text.lastIndexOf("\n", start - 1) + 1;
  const head = text.slice(lineStart, start);
  const match =
    /^(?<indent>\s*)(?:(?<bullet>[-*+]) (?<task>\[[ xX]\] )?|(?<number>\d+)\. |> ?)/.exec(head);
  if (match === null) {
    return null;
  }
  const [marker] = match;
  const { indent = "", bullet, task, number } = match.groups ?? {};
  const rest = text.slice(start);
  const isEmptyItem = head === marker && /^(?:\n|$)/.test(rest);
  if (isEmptyItem) {
    return {
      selection: { end: lineStart, start: lineStart },
      text: text.slice(0, lineStart) + text.slice(start),
    };
  }
  const nextMarker =
    number === undefined
      ? `${indent}${bullet === undefined ? "> " : `${bullet} ${task === undefined ? "" : "[ ] "}`}`
      : `${indent}${Number(number) + 1}. `;
  const inserted = `\n${nextMarker}`;
  const cursor = start + inserted.length;
  return {
    selection: { end: cursor, start: cursor },
    text: text.slice(0, start) + inserted + rest,
  };
};

type EnterKeyArgs = {
  event: {
    key: string;
    shiftKey: boolean;
    nativeEvent: { isComposing: boolean };
    preventDefault: () => void;
  };
  text: string;
  textarea: HTMLTextAreaElement | null;
  // false なら Enter も Shift+Enter と同じく改行にする
  submitsOnEnter: boolean;
  onSubmit: () => void;
  onReplace: (next: { text: string; selection: Selection }) => void;
};

// Enter で送信し、Shift+Enter はリストの記号を引き継いで改行する
export const handleEnterKey = ({
  event,
  text,
  textarea,
  submitsOnEnter,
  onSubmit,
  onReplace,
}: EnterKeyArgs) => {
  if (event.key !== "Enter" || event.nativeEvent.isComposing) {
    return;
  }
  if (submitsOnEnter && !event.shiftKey) {
    event.preventDefault();
    onSubmit();
    return;
  }
  const continued =
    textarea && continueList(text, { end: textarea.selectionEnd, start: textarea.selectionStart });
  if (continued) {
    event.preventDefault();
    onReplace(continued);
  }
};
