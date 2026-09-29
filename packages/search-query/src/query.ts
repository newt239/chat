export const searchHasValues = ["image", "file", "link", "video", "location"] as const;
export type SearchHas = (typeof searchHasValues)[number];

export const searchIsValues = ["pinned", "thread", "mention"] as const;
export type SearchIs = (typeof searchIsValues)[number];

// 修飾子を取り除いた語と、名前のままの絞り込み条件。名前から ID への解決は利用側で行う
export type SearchQuery = {
  keywords: string[];
  from: string[];
  in: string[];
  has: SearchHas[];
  is: SearchIs[];
  // YYYY-MM-DD。どちらも指定した日を含む
  after: string | null;
  before: string | null;
  // after: / before: の値が YYYY-MM-DD でない修飾子
  invalidDates: string[];
};

export const emptySearchQuery: SearchQuery = {
  after: null,
  before: null,
  from: [],
  has: [],
  in: [],
  invalidDates: [],
  is: [],
  keywords: [],
};

const datePattern = /^\d{4}-\d{2}-\d{2}$/;

const pick = <T extends string>(values: readonly T[], value: string) =>
  values.find((candidate) => candidate === value.toLowerCase());

const addUnique = <T>(list: T[], value: T) => (list.includes(value) ? list : [...list, value]);

// 空白で区切る。ダブルクォートの中の空白は区切りにしない
const tokenize = (raw: string) => raw.match(/(?:[^\s"]+|"[^"]*"?)+/g) ?? [];

// @"Alice Smith" のように記号の後ろのクォートも外す
const unquote = (value: string) => value.replace(/^([@#]?)"(.*?)"?$/, "$1$2");

// 解釈できない修飾子（値が空・未知の値）は語として残す。日付の形式が不正なものは invalidDates に分ける
export const parseSearchQuery = (raw: string) =>
  tokenize(raw).reduce<SearchQuery>((query, token) => {
    const match = /^(from|in|has|is|before|after):(.+)$/i.exec(token);
    const key = match?.[1]?.toLowerCase();
    const value = unquote(match?.[2] ?? "");
    const keep = { ...query, keywords: [...query.keywords, token] };
    const invalid = { ...query, invalidDates: addUnique(query.invalidDates, token) };
    switch (key) {
      case "from": {
        const name = value.replace(/^@/, "");
        return name ? { ...query, from: addUnique(query.from, name) } : keep;
      }
      case "in": {
        const name = value.replace(/^#/, "");
        return name ? { ...query, in: addUnique(query.in, name) } : keep;
      }
      case "has": {
        const has = pick(searchHasValues, value);
        return has ? { ...query, has: addUnique(query.has, has) } : keep;
      }
      case "is": {
        const is = pick(searchIsValues, value);
        return is ? { ...query, is: addUnique(query.is, is) } : keep;
      }
      case "before":
        return isValidDate(value) ? { ...query, before: value } : invalid;
      case "after":
        return isValidDate(value) ? { ...query, after: value } : invalid;
      default:
        return keep;
    }
  }, emptySearchQuery);

const quote = (value: string) => (/[\s"]/.test(value) ? `"${value.replaceAll('"', "")}"` : value);

export const formatSearchQuery = (query: SearchQuery) =>
  [
    ...query.keywords,
    ...query.invalidDates,
    ...query.from.map((name) => `from:@${quote(name)}`),
    ...query.in.map((name) => `in:#${quote(name)}`),
    ...query.has.map((has) => `has:${has}`),
    ...query.is.map((is) => `is:${is}`),
    query.after && `after:${query.after}`,
    query.before && `before:${query.before}`,
  ]
    .filter(Boolean)
    .join(" ");

export const hasSearchConditions = (query: SearchQuery) =>
  query.from.length > 0 ||
  query.in.length > 0 ||
  query.has.length > 0 ||
  query.is.length > 0 ||
  query.after !== null ||
  query.before !== null;

const parseDate = (value: string) => {
  const [year = 0, month = 1, day = 1] = value.split("-").map(Number);
  return new Date(year, month - 1, day);
};

// 2026-02-30 のような存在しない日付を弾く
const isValidDate = (value: string) => {
  if (!datePattern.test(value)) {
    return false;
  }
  const date = parseDate(value);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` === value;
};

// 端末のローカル日付で日時の範囲にする。返す before は指定日の翌日 0 時で、その時刻を含まない
export const searchDateRange = (query: SearchQuery) => {
  const before = query.before === null ? undefined : parseDate(query.before);
  return {
    after: query.after === null ? undefined : parseDate(query.after),
    before: before && new Date(before.getFullYear(), before.getMonth(), before.getDate() + 1),
  };
};
