const MAX_DEPTH = 4;
const MAX_SEGMENT_LENGTH = 32;
const SEGMENT_PATTERN = /^[a-z0-9_-]+$/;

// "dev/frontend/web" -> ["dev", "dev/frontend"]
export const ancestorPaths = (path: string) => {
  const segments = path.split("/");
  return segments.slice(0, -1).map((_, index) => segments.slice(0, index + 1).join("/"));
};

export const parentPath = (path: string) => ancestorPaths(path).at(-1) ?? null;

export const canHaveChildChannel = (path: string) => path.split("/").length < MAX_DEPTH;

export const lastSegment = (path: string) => path.split("/").at(-1) ?? path;

// サーバーと同じ制約（小文字の英数字・ハイフン・アンダースコア、各階層 32 文字以内・4 階層まで）
export const validateChannelPath = (path: string, existingNames: readonly string[]) => {
  if (path.length === 0) {
    return "required";
  }
  const segments = path.split("/");
  if (segments.some((segment) => segment.length === 0)) {
    return "segmentEmpty";
  }
  if (segments.length > MAX_DEPTH) {
    return "tooDeep";
  }
  if (segments.some((segment) => !SEGMENT_PATTERN.test(segment))) {
    return "invalid";
  }
  if (segments.some((segment) => segment.length > MAX_SEGMENT_LENGTH)) {
    return "segmentTooLong";
  }
  if (existingNames.includes(path)) {
    return "exists";
  }
  return null;
};
