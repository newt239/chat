import { useCallback, useEffect, useRef } from "react";

import { useMutation, useQuery } from "@connectrpc/connect-query";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";

import { useInvalidateDrafts } from "./useDrafts";

const SAVE_DELAY_MS = 800;

// 入力欄の本文を少し待ってからサーバーの下書きに保存する。チャンネルを離れるときは待たずに保存する
export const useDraftAutosave = (channelId: string, parentId: string | null) => {
  const invalidateDrafts = useInvalidateDrafts();
  const target = { channelId, parentId: parentId ?? undefined };
  const { data, isFetching } = useQuery(DraftService.method.getDraft, target, {
    refetchOnWindowFocus: false,
  });
  const { mutate: saveDraft } = useMutation(DraftService.method.saveDraft, {
    onSuccess: invalidateDrafts,
  });
  const { mutate: deleteDraft } = useMutation(DraftService.method.deleteDraft, {
    onSuccess: invalidateDrafts,
  });
  const pendingRef = useRef<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const flush = useCallback(() => {
    clearTimeout(timerRef.current);
    const body = pendingRef.current;
    if (body !== null) {
      pendingRef.current = null;
      saveDraft({ body, channelId, parentId: parentId ?? undefined });
    }
  }, [saveDraft, channelId, parentId]);

  useEffect(() => flush, [flush]);

  const save = (body: string) => {
    pendingRef.current = body;
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(flush, SAVE_DELAY_MS);
  };

  // 送信したら書きかけを消す
  const discard = () => {
    clearTimeout(timerRef.current);
    pendingRef.current = null;
    deleteDraft(target);
  };

  // キャッシュが古いまま復元しないよう、取り直しが終わってから本文を渡す
  const initialBody = data !== undefined && !isFetching ? (data.draft?.body ?? "") : null;

  return { discard, initialBody, save };
};
