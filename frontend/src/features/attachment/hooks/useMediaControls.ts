import { useRef } from "react";

import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { usePlayerState } from "#/features/player/hooks/usePlayerState";
import { mediaPlayer } from "#/features/player/mediaPlayer";

import { useAttachmentUrl, useDownloadUrl } from "../api/client";

import type { Message, MessageAttachment } from "#/gen/chat/v1/message_pb";

// メッセージ内の音声・動画の操作。再生そのものはアプリ全体で共有するプレイヤーに任せる
export const useMediaControls = (
  attachment: MessageAttachment,
  message: Message,
  kind: "audio" | "video",
) => {
  const { t } = useTranslation();
  const { workspaceId = "" } = useParams({ strict: false });
  // 再生中の添付だけが位置の更新を受け取り、ほかの添付は描き直さない
  const state = usePlayerState((current) =>
    current.track?.attachmentId === attachment.id && current.track.messageId === message.id
      ? current
      : null,
  );
  const { mutateAsync: fetchUrl } = useDownloadUrl();
  const { data: posterUrl } = useAttachmentUrl(
    kind === "video" && attachment.media?.thumbnail ? attachment.id : null,
    true,
  );
  const isActive = state !== null;

  const start = async () => {
    // React Compiler が try/catch の中の ?? を扱えないため、外で組み立てる
    const track = {
      attachmentId: attachment.id,
      authorName: message.user?.displayName ?? "",
      channelId: message.channelId,
      durationSeconds: attachment.media?.durationSeconds ?? 0,
      fileName: attachment.fileName,
      kind,
      messageId: message.id,
      parentId: message.parentId,
      posterUrl,
      workspaceId,
    };
    try {
      await mediaPlayer.play(track, async () => {
        const { url } = await fetchUrl({ attachmentId: attachment.id });
        return url;
      });
    } catch {
      toast(t("attachment.player.playFailed"), { tone: "danger" });
    }
  };

  const toggle = () => {
    if (!isActive) {
      void start();
    } else if (state.isPlaying) {
      mediaPlayer.pause();
    } else {
      void mediaPlayer.resume();
    }
  };

  const seek = (position: number) => {
    if (isActive) {
      mediaPlayer.seek(position);
      return;
    }
    void start().then(() => {
      mediaPlayer.seek(position);
    });
  };

  // 再生中の添付がメッセージ一覧に見えているかをプレイヤーに伝える
  const detachRef = useRef<(() => void) | null>(null);
  const inlineRef = (element: HTMLElement | null) => {
    detachRef.current?.();
    detachRef.current = element !== null && isActive ? mediaPlayer.attachInline(element) : null;
  };

  return {
    duration: state?.duration ?? attachment.media?.durationSeconds ?? 0,
    handleCycleRate: isActive ? mediaPlayer.cycleRate : toggle,
    handleSeek: seek,
    handleToggle: toggle,
    inlineRef,
    isActive,
    isPlaying: state?.isPlaying ?? false,
    position: state?.position ?? 0,
    rate: state?.rate ?? 1,
  };
};
