// 共有用のリンクはアプリの中ではなく Web 版の URL にする
export const toShareUrl = (href: string) => new URL(href, globalThis.location.origin).href;
