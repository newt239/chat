// 順序が意味を持つ（Edge や Chrome の UA には Safari も含まれる）
const browsers = [
  ["Edge", /Edg\//],
  ["Opera", /OPR\//],
  ["Firefox", /Firefox\//],
  ["Chrome", /Chrome\//],
  ["Safari", /Safari\//],
] as const;

const systems = [
  ["iOS", /iPhone|iPad/],
  ["Android", /Android/],
  ["Windows", /Windows/],
  ["macOS", /Mac OS X/],
  ["Linux", /Linux/],
] as const;

// "Chrome · macOS" のように端末を短く表す。判別できなければ null
export const summarizeUserAgent = (userAgent: string) => {
  const parts = [browsers, systems]
    .map((candidates) => candidates.find(([, pattern]) => pattern.test(userAgent))?.[0])
    .filter((part) => part !== undefined);
  return parts.length === 0 ? null : parts.join(" · ");
};
