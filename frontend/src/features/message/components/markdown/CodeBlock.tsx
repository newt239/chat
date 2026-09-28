import { isValidElement } from "react";
import type { ReactNode } from "react";

import { useQuery } from "@tanstack/react-query";
import { toJsxRuntime } from "hast-util-to-jsx-runtime";
import { useTranslation } from "react-i18next";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";

import { Button } from "#/components/ui/Button";
import { toast } from "#/components/ui/toast";
import { highlightCode } from "#/features/message/utils/highlight";

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
  const { data: highlighted } = useQuery({
    queryFn: () => highlightCode(code, language),
    queryKey: ["highlight", language, code],
    staleTime: Number.POSITIVE_INFINITY,
  });

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      toast(t("codeBlock.copied"), { tone: "success" });
    } catch {
      toast(t("codeBlock.copyFailed"), { tone: "danger" });
    }
  };

  return (
    <div className="my-1 max-w-[660px] overflow-hidden rounded-md border border-border bg-sunken">
      <div className="flex items-center gap-2 border-b border-border py-[3px] pr-1 pl-2.5 font-mono text-[11px] text-muted">
        <span>{language}</span>
        <span className="flex-1 text-subtle">
          {t("codeBlock.lines", { count: code.split("\n").length })}
        </span>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 px-2 font-sans text-[11px]"
          onPress={() => {
            void copy();
          }}
        >
          {t("codeBlock.copy")}
        </Button>
      </div>
      <div className="overflow-x-auto px-3 py-2 font-mono text-mono text-text [&_pre]:m-0 [&_pre]:bg-transparent!">
        {highlighted ? (
          toJsxRuntime(highlighted, { Fragment, jsx, jsxs })
        ) : (
          <pre>
            <code>{code}</code>
          </pre>
        )}
      </div>
    </div>
  );
};
