export type SuggestionQuery = {
  trigger: "@" | "#";
  query: string;
  // トリガー文字の位置
  start: number;
};

export type SuggestionItem = {
  id: string;
  kind: "user" | "group" | "channel";
  label: string;
  // 本文に挿入する文字列（トリガー文字を含む）
  value: string;
  avatarUrl: string | undefined;
};

// 行頭か空白の直後に打った @ / # から、カーソルまでを検索語にする
const tokenPattern = /(?:^|\s)(?<trigger>[@#])(?<query>[\w/-]*)$/;

export const findSuggestionQuery = (text: string, cursor: number): SuggestionQuery | null => {
  const match = tokenPattern.exec(text.slice(0, cursor));
  if (match === null) {
    return null;
  }
  const query = match.groups?.query ?? "";
  return {
    query,
    start: cursor - query.length - 1,
    trigger: match.groups?.trigger === "#" ? "#" : "@",
  };
};

// 検索語を候補に置き換え、続けて打てるよう空白を足す
export const applySuggestion = ({
  text,
  query: { start },
  cursor,
  value,
}: {
  text: string;
  query: SuggestionQuery;
  cursor: number;
  value: string;
}) => {
  const inserted = `${value} `;
  return {
    cursor: start + inserted.length,
    text: `${text.slice(0, start)}${inserted}${text.slice(cursor)}`,
  };
};

// メンションは表示名の先頭の英数字で解決されるため、それを挿入する。英数字で始まらない名前はメンションできない
export const mentionTokenOf = (displayName: string) => /^[\w-]+/.exec(displayName)?.[0] ?? null;

// 前方一致を先に、部分一致をその後に並べる
export const rankByQuery = <T>(
  candidates: readonly T[],
  query: string,
  keyOf: (candidate: T) => string,
) => {
  const lower = query.toLowerCase();
  const prefix: T[] = [];
  const partial: T[] = [];
  for (const candidate of candidates) {
    const key = keyOf(candidate).toLowerCase();
    if (key.startsWith(lower)) {
      prefix.push(candidate);
    } else if (key.includes(lower)) {
      partial.push(candidate);
    }
  }
  return [...prefix, ...partial];
};
