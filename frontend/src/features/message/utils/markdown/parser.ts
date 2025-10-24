import remarkGfm from 'remark-gfm';
import remarkParse from 'remark-parse';
import { unified } from 'unified';

import type { Root } from 'mdast';

export interface ParsedMarkdown {
  ast: Root;
}

export function parseMarkdown(content: string): ParsedMarkdown {
  const processor = unified()
    .use(remarkParse)
    .use(remarkGfm);

  const ast = processor.parse(content) as Root;
  return { ast };
}
