// 名前は後から変わるため、本文にはメンションとチャンネルを ID で埋め込む。バックエンドの entity/mention.go と同じ記法
export type MentionToken =
  | { kind: "user"; id: string }
  | { kind: "group"; id: string }
  | { kind: "channel"; id: string }
  | { kind: "broadcast"; id: "channel" | "here" };

const uuid = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}";
const mentionTokenPattern = new RegExp(
  `<(?:@(?<broadcast>channel|here)|@&(?<group>${uuid})|@(?<user>${uuid})|#(?<channel>${uuid}))>`,
  "g",
);

const parseMentionToken = (groups: Record<string, string | undefined>) => {
  if (groups.broadcast === "channel" || groups.broadcast === "here") {
    return { id: groups.broadcast, kind: "broadcast" } satisfies MentionToken;
  }
  if (groups.group !== undefined) {
    return { id: groups.group, kind: "group" } satisfies MentionToken;
  }
  if (groups.user !== undefined) {
    return { id: groups.user, kind: "user" } satisfies MentionToken;
  }
  if (groups.channel !== undefined) {
    return { id: groups.channel, kind: "channel" } satisfies MentionToken;
  }
  return null;
};

const prefixes = { broadcast: "@", channel: "#", group: "@&", user: "@" } satisfies Record<
  MentionToken["kind"],
  string
>;

export const toMentionToken = (token: MentionToken) => `<${prefixes[token.kind]}${token.id}>`;

export type MentionPart = { kind: "text"; text: string } | MentionToken;

// 本文を文字列と ID 記法の並びに分ける
export const splitMentionTokens = (body: string) => {
  const parts: MentionPart[] = [];
  let last = 0;
  for (const match of body.matchAll(mentionTokenPattern)) {
    const token = parseMentionToken(match.groups ?? {});
    if (token !== null) {
      parts.push({ kind: "text", text: body.slice(last, match.index) }, token);
      last = match.index + match[0].length;
    }
  }
  parts.push({ kind: "text", text: body.slice(last) });
  return parts.filter((part) => part.kind !== "text" || part.text !== "");
};

// 本文の ID 記法を replace の戻り値に置き換える
export const replaceMentionTokens = (body: string, replace: (token: MentionToken) => string) =>
  splitMentionTokens(body)
    .map((part) => (part.kind === "text" ? part.text : replace(part)))
    .join("");
