import { describe, it, expect } from 'vitest';

import { parseMarkdown } from "@/features/message/utils/markdown/parser";

describe('parseMarkdown', () => {
  it('基本的なMarkdownをパースできる', () => {
    const content = '# Hello\n\nThis is **bold** text.';
    const result = parseMarkdown(content);
    expect(result.ast).toBeDefined();
    expect(result.ast.type).toBe('root');
  });

  it('コードブロックをパースできる', () => {
    const content = '```typescript\nconst x = 1;\n```';
    const result = parseMarkdown(content);
    expect(result.ast).toBeDefined();
  });

  it('リストをパースできる', () => {
    const content = '- Item 1\n- Item 2\n- Item 3';
    const result = parseMarkdown(content);
    expect(result.ast).toBeDefined();
  });
});
