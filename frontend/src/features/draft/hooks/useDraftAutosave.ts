import { useEffect, useRef } from "react";

import { useMutation, useQuery } from "@connectrpc/connect-query";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";

import { useInvalidateDrafts } from "./useDrafts";

const SAVE_DELAY_MS = 800;

type Pending = { body: string; channelId: string; parentId: string | undefined };
type Autosave = { pending: Pending | null; timer: ReturnType<typeof setTimeout> | undefined };

// 待っている本文があれば、入力したときの宛先に保存する
const flush = (autosave: Autosave, saveDraft: (pending: Pending) => void) => {
  clearTimeout(autosave.timer);
  if (autosave.pending !== null) {
    saveDraft(autosave.pending);
    autosave.pending = null;
  }
};

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
  const autosaveRef = useRef<Autosave>({ pending: null, timer: undefined });

  useEffect(() => {
    const autosave = autosaveRef.current;
    return () => {
      flush(autosave, saveDraft);
    };
  }, [saveDraft, channelId, parentId]);

  const save = (body: string) => {
    const autosave = autosaveRef.current;
    clearTimeout(autosave.timer);
    autosave.pending = { ...target, body };
    autosave.timer = setTimeout(() => {
      flush(autosave, saveDraft);
    }, SAVE_DELAY_MS);
  };

  // 送信したら書きかけを消す
  const discard = () => {
    const autosave = autosaveRef.current;
    clearTimeout(autosave.timer);
    autosave.pending = null;
    deleteDraft(target);
  };

  // キャッシュが古いまま復元しないよう、取り直しが終わってから本文を渡す
  const initialBody = data !== undefined && !isFetching ? (data.draft?.body ?? "") : null;

  return { discard, initialBody, save };
};
