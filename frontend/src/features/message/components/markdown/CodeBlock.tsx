import { isValidElement } from "react";
import type { ReactNode } from "react";

import { CodeHighlight } from "@mantine/code-highlight";

type CodeBlockProps = {
  children?: ReactNode;
  className?: string;
};

const extractTextContent = (node: ReactNode): string => {
  if (typeof node === "string") {
    return node;
  }
  if (Array.isArray(node)) {
    return node.map((child: ReactNode) => extractTextContent(child)).join("");
  }
  if (isValidElement<{ children?: ReactNode }>(node)) {
    return extractTextContent(node.props.children);
  }
  return "";
};

export const CodeBlock = ({ children, className }: CodeBlockProps) => {
  // 言語情報を取得 (例: language-typescript)
  const language = className?.replace(/^language-/, "") || "plaintext";

  // コード内容を抽出
  const code = extractTextContent(children);

  return (
    <CodeHighlight
      code={code}
      language={language}
      withCopyButton
      copyLabel="コピー"
      copiedLabel="コピーしました"
    />
  );
};
