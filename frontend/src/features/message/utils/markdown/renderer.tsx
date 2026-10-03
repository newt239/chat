import * as prod from "react/jsx-runtime";
import rehypeReact from "rehype-react";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import remarkParse from "remark-parse";
import remarkRehype from "remark-rehype";
import { unified } from "unified";

import { CodeBlock } from "#/features/message/components/markdown/CodeBlock";
import { LinkComponent } from "#/features/message/components/markdown/LinkComponent";
import { MarkdownSpan } from "#/features/message/components/markdown/MarkdownSpan";

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
      ["className", "mention", "custom-emoji"],
      "dataMention",
      "dataEmoji",
    ],
  },
};

const processor = unified()
  .use(remarkParse)
  .use(remarkGfm)
  .use(remarkHideLinks)
  .use(remarkMention)
  .use(remarkCustomEmoji)
  .use(remarkRehype)
  .use(rehypeSanitize, customSchema)
  .use(rehypeReact, {
    ...prod,
    components: {
      a: LinkComponent,
      pre: CodeBlock,
      span: MarkdownSpan,
    },
  });

// hiddenUrls のリンクは本文に出さない（引用カードで表示するメッセージへのリンクなど）
export const renderMarkdown = (content: string, hiddenUrls: readonly string[]) =>
  processor.processSync({ data: { hiddenUrls }, value: content }).result;
