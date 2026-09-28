import { OgpCard } from "./OgpCard";

import type { MessageLink } from "#/gen/chat/v1/message_pb";

type LinkPreviewEmbedProps = {
  link: MessageLink;
};

// 同じワークスペースのメッセージへのリンクは OGP がないため表示しない
export const LinkPreviewEmbed = ({ link }: LinkPreviewEmbedProps) =>
  link.ogp?.title ? <OgpCard url={link.url} ogp={link.ogp} /> : null;
