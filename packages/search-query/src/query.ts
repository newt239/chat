export const searchHasValues = ["image", "file", "link", "video"] as const;
export type SearchHas = (typeof searchHasValues)[number];

export const searchIsValues = ["pinned", "thread", "mention"] as const;
export type SearchIs = (typeof searchIsValues)[number];

export const searchDuringValues = ["today", "week", "month"] as const;
export type SearchDuring = (typeof searchDuringValues)[number];

// 修飾子を取り除いた語と、名前のままの絞り込み条件。名前から ID への解決は利用側で行う
export type SearchQuery = {
  keywords: string[];
  from: string[];
  in: string[];
  has: SearchHas[];
  is: SearchIs[];
  during: SearchDuring | null;
  // YYYY-MM-DD
  after: string | null;
  before: string | null;
};

export const emptySearchQuery: SearchQuery = {
  after: null,
  before: null,
  during: null,
  from: [],
  has: [],
  in: [],
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

// 解釈できない修飾子（値が空・未知の値）は語として残す
export const parseSearchQuery = (raw: string) =>
  tokenize(raw).reduce<SearchQuery>((query, token) => {
    const match = /^(from|in|has|is|during|before|after):(.+)$/i.exec(token);
    const key = match?.[1]?.toLowerCase();
    const value = unquote(match?.[2] ?? "");
    const keep = { ...query, keywords: [...query.keywords, token] };
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
      case "during": {
        const during = pick(searchDuringValues, value);
        return during ? { ...query, during } : keep;
      }
      case "before":
        return isValidDate(value) ? { ...query, before: value } : keep;
      case "after":
        return isValidDate(value) ? { ...query, after: value } : keep;
      default:
        return keep;
    }
  }, emptySearchQuery);

const quote = (value: string) => (/[\s"]/.test(value) ? `"${value.replaceAll('"', "")}"` : value);

export const formatSearchQuery = (query: SearchQuery) =>
  [
    ...query.keywords,
    ...query.from.map((name) => `from:@${quote(name)}`),
    ...query.in.map((name) => `in:#${quote(name)}`),
    ...query.has.map((has) => `has:${has}`),
    ...query.is.map((is) => `is:${is}`),
    query.during && `during:${query.during}`,
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
  query.during !== null ||
  query.after !== null ||
  query.before !== null;

const startOfDay = (date: Date) => new Date(date.getFullYear(), date.getMonth(), date.getDate());

const addDays = (date: Date, days: number) =>
  new Date(date.getFullYear(), date.getMonth(), date.getDate() + days);

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

const duringDays: Record<SearchDuring, number> = { month: 30, today: 1, week: 7 };

const latest = (dates: Date[]) =>
  dates.length === 0 ? undefined : new Date(Math.max(...dates.map((date) => date.getTime())));

const earliest = (dates: Date[]) =>
  dates.length === 0 ? undefined : new Date(Math.min(...dates.map((date) => date.getTime())));

// 期間を now のローカル日付で日時の範囲にする。after は含み、before は含まない
// after: / before: は Slack と同じく指定した日を含まない
export const searchDateRange = (query: SearchQuery, now: Date) => {
  const afters: Date[] = [];
  const befores: Date[] = [];
  if (query.during) {
    afters.push(addDays(startOfDay(now), 1 - duringDays[query.during]));
  }
  if (query.after) {
    afters.push(addDays(parseDate(query.after), 1));
  }
  if (query.before) {
    befores.push(parseDate(query.before));
  }
  return { after: latest(afters), before: earliest(befores) };
};
