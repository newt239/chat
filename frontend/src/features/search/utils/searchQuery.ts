export const searchHasValues = ["image", "file", "link", "video", "location"] as const;
export type SearchHas = (typeof searchHasValues)[number];

const searchIsValues = ["pinned", "thread", "mention"] as const;
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

const pad = (n: number) => String(n).padStart(2, "0");

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
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` === value;
};

const pick = <T extends string>(values: readonly T[], value: string) =>
  values.find((candidate) => candidate === value.toLowerCase());

const addUnique = <T>(list: T[], value: T | undefined) => {
  if (value === undefined || value === "") {
    return false;
  }
  if (!list.includes(value)) {
    list.push(value);
  }
  return true;
};

// 空白で区切る。ダブルクォートの中の空白は区切りにしない
const tokenize = (raw: string) => raw.match(/(?:[^\s"]+|"[^"]*"?)+/g) ?? [];

// @"Alice Smith" のように記号の後ろのクォートも外す
const unquote = (value: string) =>
  value.replace(/^(?<mark>[@#]?)"(?<body>.*?)"?$/, "$<mark>$<body>");

const modifierPattern = /^(?<key>from|in|has|is|before|after):(?<value>.+)$/i;

// 解釈できない修飾子（値が空・未知の値）は語として残す。日付の形式が不正なものは invalidDates に分ける
export const parseSearchQuery = (raw: string) => {
  const query: SearchQuery = {
    ...emptySearchQuery,
    from: [],
    has: [],
    in: [],
    invalidDates: [],
    is: [],
    keywords: [],
  };
  for (const token of tokenize(raw)) {
    const groups = modifierPattern.exec(token)?.groups;
    const key = groups?.key?.toLowerCase();
    const value = unquote(groups?.value ?? "");
    if (key === "before" || key === "after") {
      if (isValidDate(value)) {
        query[key] = value;
      } else {
        addUnique(query.invalidDates, token);
      }
      continue;
    }
    const isApplied =
      (key === "from" && addUnique(query.from, value.replace(/^@/, ""))) ||
      (key === "in" && addUnique(query.in, value.replace(/^#/, ""))) ||
      (key === "has" && addUnique(query.has, pick(searchHasValues, value))) ||
      (key === "is" && addUnique(query.is, pick(searchIsValues, value)));
    if (!isApplied) {
      query.keywords.push(token);
    }
  }
  return query;
};

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

// 端末のローカル日付で日時の範囲にする。返す before は指定日の翌日 0 時で、その時刻を含まない
export const searchDateRange = (query: SearchQuery) => {
  const before = query.before === null ? undefined : parseDate(query.before);
  return {
    after: query.after === null ? undefined : parseDate(query.after),
    before: before && new Date(before.getFullYear(), before.getMonth(), before.getDate() + 1),
  };
};
