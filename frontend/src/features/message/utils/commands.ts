// サーバーで実行できるコマンド。使い方は辞書の command.<名前>.usage に置く
export const commandNames = ["remind"] as const;

// 入力が既知のコマンドならその名前を返す。「//」で始めるとコマンドにせず「/」から始まる文章として送る
export const findCommand = (body: string) => {
  const name = /^\/(?<name>[a-z]+)(?:\s|$)/.exec(body)?.groups?.name;
  return commandNames.find((candidate) => candidate === name) ?? null;
};

export const unescapeCommand = (body: string) => (body.startsWith("//") ? body.slice(1) : body);
