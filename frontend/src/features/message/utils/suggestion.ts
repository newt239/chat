type SuggestionQuery = {
  trigger: "@" | "#" | "/";
  query: string;
  // トリガー文字の位置
  start: number;
};

export type SuggestionItem = {
  id: string;
  kind: "user" | "group" | "channel" | "broadcast" | "command";
  label: string;
  // 入力欄に挿入する文字列（トリガー文字を含む）
  value: string;
  // 送信するときに value を置き換える ID 記法
  token: string;
  avatarUrl: string | undefined;
};

// 行頭か空白の直後に打った @ / # から、カーソルまでを検索語にする。日本語の名前も探せるよう空白以外を受け付ける
const tokenPattern = /(?:^|\s)(?<trigger>[@#])(?<query>[^\s@#]*)$/;

// コマンドは入力欄の先頭で打ったものだけ
const commandPattern = /^\/(?<query>[a-z]*)$/;

export const findSuggestionQuery = (
  text: string,
  cursor: number,
  allowsCommands: boolean,
): SuggestionQuery | null => {
  const command = allowsCommands ? commandPattern.exec(text.slice(0, cursor)) : null;
  if (command !== null) {
    return { query: command.groups?.query ?? "", start: 0, trigger: "/" };
  }
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
