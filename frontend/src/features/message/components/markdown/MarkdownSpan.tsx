import type { ReactNode } from "react";

import { CustomEmoji } from "#/features/customEmoji/components/CustomEmoji";
import { Mention } from "#/features/mention/components/Mention";
import { splitMentionTokens } from "#/features/mention/utils/mentionToken";

type MarkdownSpanProps = {
  className?: string;
  "data-mention"?: string;
  "data-emoji"?: string;
  children?: ReactNode;
};

// 本文の span をメンション・チャンネル・カスタム絵文字に置き換える。毎回作ると再描画で押下中のボタンが作り直される
export const MarkdownSpan = (props: MarkdownSpanProps) => {
  const classNames = props.className?.split(" ") ?? [];
  const [token] = classNames.includes("mention")
    ? splitMentionTokens(props["data-mention"] ?? "")
    : [];
  if (token && token.kind !== "text") {
    return <Mention token={token} />;
  }
  if (classNames.includes("custom-emoji") && props["data-emoji"]) {
    return <CustomEmoji name={props["data-emoji"]} />;
  }
  return <span {...props} />;
};
