import { MessageQuoteCard } from "./MessageQuoteCard";
import { OgpCard } from "./OgpCard";
import { XPostCard } from "./XPostCard";
import { YouTubeCard } from "./YouTubeCard";

import type { MessageLink } from "#/gen/chat/v1/message_pb";

type LinkPreviewEmbedProps = {
  link: MessageLink;
};

// 同じワークスペースのメッセージは引用、YouTube は動画、X は投稿のカード、それ以外は OGP のカードにする
export const LinkPreviewEmbed = ({ link }: LinkPreviewEmbedProps) => {
  if (link.linkedMessageId !== undefined) {
    return <MessageQuoteCard link={link} />;
  }
  if (link.ogp?.youtube !== undefined) {
    return <YouTubeCard url={link.url} ogp={link.ogp} video={link.ogp.youtube} />;
  }
  if (link.ogp?.xPost !== undefined) {
    return <XPostCard url={link.url} ogp={link.ogp} post={link.ogp.xPost} />;
  }
  return link.ogp?.title ? <OgpCard url={link.url} ogp={link.ogp} /> : null;
};
