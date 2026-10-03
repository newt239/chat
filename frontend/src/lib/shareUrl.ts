// 共有用のリンクはアプリの中ではなく Web 版の URL にする
export const toShareUrl = (href: string) =>
  new URL(href, import.meta.env.VITE_PUBLIC_APP_URL || globalThis.location.origin).href;
