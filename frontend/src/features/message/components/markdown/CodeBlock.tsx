import { isValidElement, useState } from "react";
import type { ReactNode } from "react";

import { useQuery } from "@tanstack/react-query";
import { toJsxRuntime } from "hast-util-to-jsx-runtime";
import { useTranslation } from "react-i18next";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";

import { Button } from "#/components/ui/Button/Button";
import { cn } from "#/components/ui/styles/styles";
import { highlightCode } from "#/features/message/utils/highlight";
import { copyWithToast } from "#/lib/clipboard";

// これより長いコードは折りたたみ、「すべて表示」で広げる
const COLLAPSE_LINES = 12;

type CodeBlockProps = {
  children?: ReactNode;
};

// 言語のクラス（language-ts など）は pre ではなく子の code 要素に付く
const extractLanguage = (node: ReactNode): string => {
  if (Array.isArray(node)) {
    return node.map((child: ReactNode) => extractLanguage(child)).find(Boolean) ?? "";
  }
  if (isValidElement<{ className?: string }>(node)) {
    return node.props.className?.match(/language-(?<language>\S+)/)?.groups?.language ?? "";
  }
  return "";
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

export const CodeBlock = ({ children }: CodeBlockProps) => {
  const { t } = useTranslation();
  const language = extractLanguage(children) || "text";
  const code = extractTextContent(children).replace(/\n$/, "");
  const lineCount = code.split("\n").length;
  const [isExpanded, setIsExpanded] = useState(false);
  const isCollapsed = lineCount > COLLAPSE_LINES && !isExpanded;
  const { data: highlighted } = useQuery({
    queryFn: () => highlightCode(code, language),
    queryKey: ["highlight", language, code],
    staleTime: Number.POSITIVE_INFINITY,
  });

  return (
    <div className="my-1 max-w-165 overflow-hidden rounded-md border border-border bg-sunken">
      <div className="flex items-center gap-2 border-b border-border py-0.75 pr-1 pl-2.5 font-mono text-caption text-muted">
        <span>{language}</span>
        <span className="flex-1 text-subtle">{t("codeBlock.lines", { count: lineCount })}</span>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 px-2 font-sans text-caption"
          onPress={() => {
            void copyWithToast(code, t("codeBlock.copied"));
          }}
        >
          {t("codeBlock.copy")}
        </Button>
      </div>
      <div
        className={cn(
          "overflow-x-auto px-3 py-2 font-mono text-mono text-text [&_pre]:m-0 [&_pre]:bg-transparent!",
          isCollapsed &&
            "max-h-55 overflow-hidden [mask-image:linear-gradient(black_70%,transparent)]",
        )}
      >
        {highlighted ? (
          toJsxRuntime(highlighted, { Fragment, jsx, jsxs })
        ) : (
          <pre>
            <code>{code}</code>
          </pre>
        )}
      </div>
      {isCollapsed && (
        <Button
          variant="ghost"
          size="sm"
          className="h-auto w-full rounded-none border-t border-border py-1 text-accent-text"
          onPress={() => {
            setIsExpanded(true);
          }}
        >
          <span className="text-xs font-semibold">
            {t("codeBlock.showAll", { count: lineCount })}
          </span>
        </Button>
      )}
    </div>
  );
};
