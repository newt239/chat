import type { ReactNode } from "react";

import * as prod from "react/jsx-runtime";
import rehypeReact from "rehype-react";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import { unified } from "unified";

import { remarkChannel } from "./plugins/channel";
import { remarkMention } from "./plugins/mention";

import { ChannelLink } from "@/features/message/components/markdown/ChannelLink";
import { CodeBlock } from "@/features/message/components/markdown/CodeBlock";
import { LinkComponent } from "@/features/message/components/markdown/LinkComponent";
import { Mention } from "@/features/message/components/markdown/Mention";

const customSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    code: [...(defaultSchema.attributes?.code || []), "className"],
    span: [
      ...(defaultSchema.attributes?.span || []),
      ["className", "mention", "channel-link"],
      "dataMention",
      "dataChannel",
    ],
  },
};

export function renderMarkdown(content: string): ReactNode {
  const processor = unified()
    .use(remarkParse)
    .use(remarkGfm)
    .use(remarkMention)
    .use(remarkChannel)
    .use(remarkRehype)
    .use(rehypeSanitize, customSchema)
    .use(rehypeReact, {
      ...prod,
      components: {
        pre: CodeBlock,
        a: LinkComponent,
        span: (props: {
          className?: string;
          "data-mention"?: string;
          "data-channel"?: string;
          children?: ReactNode;
        }) => {
          const classNames = props.className?.split(" ") || [];
          if (classNames.includes("mention")) {
            return <Mention {...props} data-mention={props["data-mention"] || ""} />;
          }
          if (classNames.includes("channel-link")) {
            return <ChannelLink {...props} data-channel={props["data-channel"] || ""} />;
          }
          return <span {...props} />;
        },
      },
    });

  return processor.processSync(content).result;
}
