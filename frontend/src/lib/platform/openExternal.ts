export const openExternal = (url: string) => {
  globalThis.open(url, "_blank", "noopener,noreferrer");
};
