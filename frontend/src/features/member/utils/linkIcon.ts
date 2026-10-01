import {
  IconBrandBluesky,
  IconBrandDiscord,
  IconBrandFacebook,
  IconBrandGithub,
  IconBrandInstagram,
  IconBrandLinkedin,
  IconBrandThreads,
  IconBrandTiktok,
  IconBrandTwitch,
  IconBrandX,
  IconBrandYoutube,
  IconLink,
} from "@tabler/icons-react";

import type { Icon } from "@tabler/icons-react";

// ホスト名（www. を除く）ごとのアイコン。サブドメインも同じサイトとして扱う
const siteIcons: [string, Icon][] = [
  ["x.com", IconBrandX],
  ["twitter.com", IconBrandX],
  ["instagram.com", IconBrandInstagram],
  ["youtube.com", IconBrandYoutube],
  ["youtu.be", IconBrandYoutube],
  ["github.com", IconBrandGithub],
  ["facebook.com", IconBrandFacebook],
  ["linkedin.com", IconBrandLinkedin],
  ["tiktok.com", IconBrandTiktok],
  ["bsky.app", IconBrandBluesky],
  ["threads.net", IconBrandThreads],
  ["threads.com", IconBrandThreads],
  ["discord.gg", IconBrandDiscord],
  ["discord.com", IconBrandDiscord],
  ["twitch.tv", IconBrandTwitch],
];

/** プロフィールのリンクの URL から、主要なサイトならそのアイコン、それ以外は汎用のリンクのアイコンを返す */
export const linkIconOf = (url: string) => {
  const host = URL.canParse(url) ? new URL(url).hostname.replace(/^www\./, "") : "";
  return (
    siteIcons.find(([domain]) => host === domain || host.endsWith(`.${domain}`))?.[1] ?? IconLink
  );
};

// 表示用に、スキーム・www.・末尾のスラッシュを除いた URL
export const displayUrl = (url: string) =>
  url.replace(/^https?:\/\/(?:www\.)?/, "").replace(/\/$/, "");
