import type { ReactNode } from "react";

import { CustomEmoji } from "#/features/customEmoji/components/CustomEmoji";
import { ChannelLink } from "#/features/mention/components/ChannelLink";
import { Mention } from "#/features/mention/components/Mention";

type MarkdownSpanProps = {
  className?: string;
  "data-mention"?: string;
  "data-channel"?: string;
  "data-emoji"?: string;
  children?: ReactNode;
};

// 本文の span をメンション・チャンネル・カスタム絵文字に置き換える。毎回作ると再描画で押下中のボタンが作り直される
export const MarkdownSpan = (props: MarkdownSpanProps) => {
  const classNames = props.className?.split(" ") ?? [];
  if (classNames.includes("mention")) {
    return <Mention {...props} data-mention={props["data-mention"] || ""} />;
  }
  if (classNames.includes("channel-link")) {
    return <ChannelLink {...props} data-channel={props["data-channel"] || ""} />;
  }
  if (classNames.includes("custom-emoji") && props["data-emoji"]) {
    return <CustomEmoji name={props["data-emoji"]} />;
  }
  return <span {...props} />;
};
