import type { ReactNode, ReactElement } from 'react';

import { Code } from '@mantine/core';

interface CodeBlockProps {
  children?: ReactNode;
  className?: string;
}

export const CodeBlock = ({ children, className }: CodeBlockProps) => {
  // 言語情報を取得 (例: language-typescript)
  const language = className?.replace(/^language-/, '') || 'text';

  // コード内容を抽出
  const code = extractTextContent(children);

  return (
    <Code block className={`language-${language}`}>
      {code}
    </Code>
  );
}

function extractTextContent(node: ReactNode): string {
  if (typeof node === 'string') {
    return node;
  }
  if (Array.isArray(node)) {
    return node.map(extractTextContent).join('');
  }
  if (node && typeof node === 'object' && 'props' in node) {
    const element = node as ReactElement;
    const props = element.props as { children?: ReactNode };
    return extractTextContent(props.children);
  }
  return '';
}
