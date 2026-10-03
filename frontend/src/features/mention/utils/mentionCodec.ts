import { replaceMentionTokens, toMentionToken } from "./mentionToken";

import type { MentionToken } from "./mentionToken";

// 入力欄に出す名前と、本文に埋め込む ID 記法の対応
export type MentionLabels = Map<string, string>;

const broadcastLabels = new Map([
  ["@channel", toMentionToken({ id: "channel", kind: "broadcast" })],
  ["@here", toMentionToken({ id: "here", kind: "broadcast" })],
]);

const escapeRegExp = (text: string) => text.replaceAll(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`);

// 入力欄の文章の名前を ID 記法に戻す。名前の続きに英数字が続くときは別の語として扱う
export const encodeMentions = (text: string, labels: MentionLabels) => {
  const all = new Map([...broadcastLabels, ...labels]);
  // 長い名前を先に試し、「@Bob Smith」を「@Bob」と取り違えない
  const sorted = [...all.keys()]
    .toSorted((a, b) => b.length - a.length)
    .map((label) => escapeRegExp(label));
  const pattern = new RegExp(`(?:${sorted.join("|")})(?![\\w-])`, "g");
  return text.replaceAll(pattern, (label) => all.get(label) ?? label);
};

// 本文の ID 記法を入力欄に出す名前にし、対応を labels に足す。名前が分からないものは記法のまま残す
export const decodeMentions = (
  body: string,
  labels: MentionLabels,
  labelOf: (token: MentionToken) => string | null,
) =>
  replaceMentionTokens(body, (token) => {
    const raw = toMentionToken(token);
    const label = labelOf(token);
    if (label === null) {
      return raw;
    }
    labels.set(label, raw);
    return label;
  });
