import type { ReactNode } from "react";

import * as prod from "react/jsx-runtime";
import rehypeReact from "rehype-react";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import { unified } from "unified";

import { CustomEmoji } from "#/features/customEmoji/components/CustomEmoji";
import { ChannelLink } from "#/features/message/components/markdown/ChannelLink";
import { CodeBlock } from "#/features/message/components/markdown/CodeBlock";
import { LinkComponent } from "#/features/message/components/markdown/LinkComponent";
import { Mention } from "#/features/message/components/markdown/Mention";

import { remarkCustomEmoji } from "./plugins/customEmoji";
import { remarkHideLinks } from "./plugins/hideLinks";
import { remarkMention } from "./plugins/mention";

const customSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    code: [...(defaultSchema.attributes?.code ?? []), "className"],
    span: [
      ...(defaultSchema.attributes?.span ?? []),
      ["className", "mention", "channel-link", "custom-emoji"],
      "dataMention",
      "dataChannel",
      "dataEmoji",
    ],
  },
};

// hiddenUrls のリンクは本文に出さない（引用カードで表示するメッセージへのリンクなど）
export const renderMarkdown = (content: string, hiddenUrls: readonly string[]): ReactNode => {
  const processor = unified()
    .use(remarkParse)
    .use(remarkGfm)
    .use(remarkHideLinks, hiddenUrls)
    .use(remarkMention)
    .use(remarkCustomEmoji)
    .use(remarkRehype)
    .use(rehypeSanitize, customSchema)
    .use(rehypeReact, {
      ...prod,
      components: {
        a: LinkComponent,
        pre: CodeBlock,
        span: (props: {
          className?: string;
          "data-mention"?: string;
          "data-channel"?: string;
          "data-emoji"?: string;
          children?: ReactNode;
        }) => {
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
        },
      },
    });

  return processor.processSync(content).result;
};
