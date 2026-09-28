import { createHighlighterCore } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";

import type { HighlighterCore, LanguageInput } from "shiki/core";

// チャットでよく貼られる言語に絞り、PWA のプリキャッシュを膨らませない
const typescript = () => import("shiki/langs/typescript.mjs");
const javascript = () => import("shiki/langs/javascript.mjs");
const shell = () => import("shiki/langs/shellscript.mjs");
const languages: Record<string, LanguageInput> = {
  bash: shell,
  css: () => import("shiki/langs/css.mjs"),
  diff: () => import("shiki/langs/diff.mjs"),
  dockerfile: () => import("shiki/langs/dockerfile.mjs"),
  go: () => import("shiki/langs/go.mjs"),
  html: () => import("shiki/langs/html.mjs"),
  java: () => import("shiki/langs/java.mjs"),
  javascript,
  js: javascript,
  json: () => import("shiki/langs/json.mjs"),
  jsx: () => import("shiki/langs/jsx.mjs"),
  kotlin: () => import("shiki/langs/kotlin.mjs"),
  markdown: () => import("shiki/langs/markdown.mjs"),
  proto: () => import("shiki/langs/proto.mjs"),
  python: () => import("shiki/langs/python.mjs"),
  rust: () => import("shiki/langs/rust.mjs"),
  sh: shell,
  shell,
  sql: () => import("shiki/langs/sql.mjs"),
  swift: () => import("shiki/langs/swift.mjs"),
  ts: typescript,
  tsx: () => import("shiki/langs/tsx.mjs"),
  typescript,
  yaml: () => import("shiki/langs/yaml.mjs"),
  yml: () => import("shiki/langs/yaml.mjs"),
};

let highlighter: Promise<HighlighterCore> | null = null;

// WASM を読み込まずに済むよう JavaScript の正規表現エンジンを使い、言語は必要になった時点で読み込む
const getHighlighter = () => {
  highlighter ??= createHighlighterCore({
    engine: createJavaScriptRegexEngine(),
    langs: [],
    themes: [import("shiki/themes/github-light.mjs"), import("shiki/themes/github-dark.mjs")],
  });
  return highlighter;
};

// 色は --shiki-light / --shiki-dark の CSS 変数で出力し、globals.css で表示モードに応じて切り替える
export const highlightCode = async (code: string, language: string) => {
  const instance = await getHighlighter();
  const loader = languages[language];
  const loaded = loader
    ? await instance.loadLanguage(loader).then(() => instance.getLoadedLanguages())
    : [];
  return instance.codeToHast(code, {
    defaultColor: false,
    lang: loaded.includes(language) ? language : "text",
    themes: { dark: "github-dark", light: "github-light" },
  });
};
