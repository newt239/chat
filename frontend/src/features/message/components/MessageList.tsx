import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import type { ReactNode } from "react";

import { IconArrowDown, IconLoader2 } from "@tabler/icons-react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";

import { useHighlightedMessage } from "../hooks/useHighlightedMessage";
import { findRowIndex } from "../utils/timelineRows";
import { DateDivider } from "./DateDivider";
import { SystemMessageItem } from "./SystemMessageItem";

import type { TimelineRow } from "../utils/timelineRows";

import type { Message } from "#/gen/chat/v1/message_pb";

type Direction = "older" | "newer";

type MessageListProps = {
  rows: TimelineRow[];
  currentUserId: string | null;
  // 開いたときにスクロールしてハイライトするメッセージ
  targetMessageId: string | null;
  hasOlder: boolean;
  hasNewer: boolean;
  loading: Direction | null;
  onLoad: (direction: Direction) => void;
  onJumpToLatest: () => void;
  // 最新のユーザーメッセージの要素を受け取り、画面に入ったら既読にする
  latestMessageRef: (element: HTMLElement | null) => void;
  latestUserMessageId: string | null;
  renderMessage: (message: Message, isHighlighted: boolean) => ReactNode;
  // kind が header の行に描画する内容
  header: ReactNode;
};

// 最下部からこの距離以内なら、新着が届いたときに最下部へ追従する
const STICK_TO_BOTTOM_PX = 80;
// 端からこの行数以内に来たら続きを読み込む
const LOAD_THRESHOLD_ROWS = 3;

const isMessageRow = (row: TimelineRow | undefined) =>
  row?.kind === "user" || row?.kind === "system";

const firstMessageKey = (rows: readonly TimelineRow[]) =>
  rows.find((row) => isMessageRow(row))?.key;

/** タイムラインを画面に見えている行だけ描画する。最下部に追従し、端に近づくと前後を読み込む */
export const MessageList = ({
  rows,
  currentUserId,
  targetMessageId,
  hasOlder,
  hasNewer,
  loading,
  onLoad,
  onJumpToLatest,
  latestMessageRef,
  latestUserMessageId,
  renderMessage,
  header,
}: MessageListProps) => {
  const { t } = useTranslation();
  const scrollRef = useRef<HTMLDivElement>(null);
  // TanStack Virtual は React Compiler と併用できないため、このコンポーネントはメモ化されない
  // oxlint-disable-next-line react/incompatible-library
  const virtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: (index) => (isMessageRow(rows[index]) ? 64 : 40),
    getItemKey: (index) => rows[index]?.key ?? index,
    getScrollElement: () => scrollRef.current,
    overscan: 8,
    // ホバー時のツールバーがメッセージの上にはみ出すため、先頭に余白を取る
    paddingEnd: 8,
    paddingStart: 32,
  });

  const scrollToMessage = useCallback(
    (messageId: string) => {
      const index = findRowIndex(rows, messageId);
      if (index === -1) {
        return false;
      }
      virtualizer.scrollToIndex(index, { align: "center" });
      return true;
    },
    [rows, virtualizer],
  );
  const highlightedId = useHighlightedMessage(rows.length > 0, targetMessageId, scrollToMessage);

  const isAtBottomRef = useRef(targetMessageId === null);
  // 最初の位置へスクロールし終えるまでは、先頭にいても過去を読み込まない
  const [isPositioned, setIsPositioned] = useState(false);
  // 過去を読み込んで先頭に行が増えても、見ていた行が同じ位置に残るようにする
  const anchorRef = useRef<{ key: string; offset: number } | null>(null);
  const prevFirstKeyRef = useRef(firstMessageKey(rows));
  const prevLastKeyRef = useRef(rows.at(-1)?.key);

  const recordScroll = () => {
    const element = scrollRef.current;
    if (element === null) {
      return;
    }
    isAtBottomRef.current =
      element.scrollHeight - element.scrollTop - element.clientHeight < STICK_TO_BOTTOM_PX;
    const top = element.scrollTop;
    // 日付の区切りは同じ日の過去が足されると動くため、メッセージの行を基準にする
    const anchor = virtualizer
      .getVirtualItems()
      .find((item) => item.end > top && isMessageRow(rows[item.index]));
    anchorRef.current = anchor ? { key: String(anchor.key), offset: anchor.start - top } : null;
  };

  // 開いたときに一度だけ位置を決める。メッセージを指定して開いたときは useHighlightedMessage が動かす
  const hasPositionedRef = useRef(false);
  useLayoutEffect(() => {
    if (hasPositionedRef.current) {
      return;
    }
    hasPositionedRef.current = true;
    if (targetMessageId === null && rows.length > 0) {
      virtualizer.scrollToIndex(rows.length - 1, { align: "end" });
    }
  }, [targetMessageId, rows.length, virtualizer]);
  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      setIsPositioned(true);
    });
    return () => {
      cancelAnimationFrame(frame);
    };
  }, []);

  useLayoutEffect(() => {
    const element = scrollRef.current;
    const firstKey = firstMessageKey(rows);
    const lastRow = rows.at(-1);
    const isPrepended = firstKey !== prevFirstKeyRef.current;
    const isAppended = lastRow?.key !== prevLastKeyRef.current;
    prevFirstKeyRef.current = firstKey;
    prevLastKeyRef.current = lastRow?.key;
    if (element === null) {
      return;
    }

    if (isPrepended && !isAtBottomRef.current && anchorRef.current !== null) {
      const index = rows.findIndex((row) => row.key === anchorRef.current?.key);
      const start = virtualizer.measurementsCache[index]?.start;
      if (start !== undefined) {
        const top = start - anchorRef.current.offset;
        element.scrollTop = top;
        // スクロールイベントを待たずに伝え、足した行の実測で高さが変わったときの補正をこの位置基準で行わせる
        virtualizer.scrollOffset = top;
      }
    }
    // 自分の投稿は上を読んでいても最下部へ移る
    if (isAppended && lastRow?.kind === "user" && lastRow.message.userId === currentUserId) {
      isAtBottomRef.current = true;
    }
  }, [rows, virtualizer, currentUserId]);

  // 新着や画像の読み込みで高さが変わっても、最下部にいれば最下部に留まる
  const totalSize = virtualizer.getTotalSize();
  useLayoutEffect(() => {
    const element = scrollRef.current;
    if (element !== null && isAtBottomRef.current && !hasNewer) {
      element.scrollTop = element.scrollHeight;
    }
  }, [totalSize, rows, hasNewer]);

  const virtualItems = virtualizer.getVirtualItems();
  const firstIndex = virtualItems.at(0)?.index ?? 0;
  const lastIndex = virtualItems.at(-1)?.index ?? 0;
  useEffect(() => {
    if (!isPositioned || loading !== null) {
      return;
    }
    if (hasOlder && firstIndex <= LOAD_THRESHOLD_ROWS) {
      onLoad("older");
    } else if (hasNewer && lastIndex >= rows.length - 1 - LOAD_THRESHOLD_ROWS) {
      onLoad("newer");
    }
  }, [isPositioned, loading, hasOlder, hasNewer, firstIndex, lastIndex, rows.length, onLoad]);

  // 表示中の最上段の日付を上端に重ねる。日付の区切りそのものが最上段にあるときは重ねない
  const topRow = rows[virtualizer.range?.startIndex ?? 0];

  return (
    <div className="relative flex min-h-0 flex-1 flex-col">
      {topRow !== undefined && isMessageRow(topRow) && (
        <DateDivider dateKey={topRow.dateKey} floating />
      )}
      <div
        ref={scrollRef}
        onScroll={recordScroll}
        className="flex min-h-0 flex-1 flex-col overflow-y-auto [overflow-anchor:none]"
      >
        <div className="relative mt-auto w-full shrink-0" style={{ height: totalSize }}>
          {virtualItems.map((item) => {
            const row = rows[item.index];
            if (row === undefined) {
              return null;
            }
            return (
              <div
                key={item.key}
                data-index={item.index}
                ref={virtualizer.measureElement}
                className="absolute top-0 left-0 w-full"
                style={{ transform: `translateY(${item.start}px)` }}
              >
                {row.kind === "header" && header}
                {row.kind === "date" && <DateDivider dateKey={row.dateKey} />}
                {row.kind === "user" && (
                  <div ref={row.message.id === latestUserMessageId ? latestMessageRef : undefined}>
                    {renderMessage(row.message, row.message.id === highlightedId)}
                  </div>
                )}
                {row.kind === "system" && <SystemMessageItem message={row.message} />}
              </div>
            );
          })}
        </div>
      </div>
      {loading !== null && (
        <div
          role="status"
          className={`pointer-events-none absolute inset-x-0 z-10 flex justify-center ${loading === "older" ? "top-10" : "bottom-2"}`}
        >
          <span className="inline-flex items-center gap-1.5 rounded-full border border-border bg-surface px-3 py-1 text-caption text-muted shadow-sm">
            <IconLoader2 aria-hidden className="size-3.5 animate-spin" />
            {t("message.panel.loading")}
          </span>
        </div>
      )}
      {hasNewer && loading === null && (
        <div className="pointer-events-none absolute inset-x-0 bottom-2 z-10 flex justify-center">
          <Button
            variant="secondary"
            size="sm"
            className="pointer-events-auto shadow-sm"
            onPress={onJumpToLatest}
          >
            <IconArrowDown aria-hidden className="size-3.5" />
            {t("message.panel.jumpToLatest")}
          </Button>
        </div>
      )}
    </div>
  );
};
