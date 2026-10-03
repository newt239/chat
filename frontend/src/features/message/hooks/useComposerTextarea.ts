import { useRef, useState } from "react";

import { handleEnterKey } from "../utils/format";
import { useComposerSuggestion } from "./useComposerSuggestion";

import type { Selection } from "../utils/format";

import type { KeyboardEvent } from "react-aria-components";

type Options = {
  body: string;
  onBodyChange: (body: string) => void;
  // 選んだ候補の名前と ID 記法の対応を覚える
  registerMention: (label: string, token: string) => void;
  allowsCommands: boolean;
  // false なら Enter も改行にする
  submitsOnEnter: boolean;
  onSubmit: () => void;
  // 候補を開いていないときの Esc
  onEscape: (() => void) | null;
};

/** 本文の入力欄のカーソル・@ / # / の候補・Enter の配線 */
export const useComposerTextarea = ({
  body,
  onBodyChange,
  registerMention,
  allowsCommands,
  submitsOnEnter,
  onSubmit,
  onEscape,
}: Options) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [selection, setSelection] = useState({ end: body.length, start: body.length });

  // select イベントはカーソルの移動では届かないため、使う直前に textarea から読む
  const readSelection = () => {
    const textarea = textareaRef.current;
    return textarea === null
      ? selection
      : { end: textarea.selectionEnd, start: textarea.selectionStart };
  };

  const replaceSelection = (next: { text: string; selection: Selection }) => {
    onBodyChange(next.text);
    setSelection(next.selection);
    // 値が反映されてから選択範囲を動かす
    requestAnimationFrame(() => {
      textareaRef.current?.focus();
      textareaRef.current?.setSelectionRange(next.selection.start, next.selection.end);
    });
  };

  const suggestion = useComposerSuggestion({
    allowsCommands,
    body,
    cursor: selection.start,
    onApply: (next, item) => {
      registerMention(item.value, item.token);
      replaceSelection({ selection: { end: next.cursor, start: next.cursor }, text: next.text });
    },
  });

  return {
    fieldProps: {
      onChange: (next: string) => {
        onBodyChange(next);
        // 候補の検索語はカーソルの位置で決まるため、入力のたびに読み直す
        setSelection(readSelection());
      },
      onKeyDown: (event: KeyboardEvent) => {
        if (suggestion.handleKeyDown(event)) {
          return;
        }
        if (event.key === "Escape" && onEscape !== null) {
          onEscape();
          return;
        }
        handleEnterKey({
          event,
          onReplace: replaceSelection,
          onSubmit,
          submitsOnEnter,
          text: body,
          textarea: textareaRef.current,
        });
      },
      value: body,
    },
    focus: () => {
      textareaRef.current?.focus();
      textareaRef.current?.setSelectionRange(selection.start, selection.end);
    },
    readSelection,
    replaceSelection,
    selection,
    setSelection,
    suggestion,
    textAreaProps: {
      ...suggestion.inputProps,
      onSelect: () => {
        setSelection(readSelection());
      },
      ref: textareaRef,
    },
  };
};
