import { useState, useCallback } from "react";

import { usePresignUpload } from "../api/client";
import { measureMedia } from "../utils/measureMedia";
import { validateFile } from "../utils/validator";

import type { PendingAttachment } from "../api/types";

type UploadOptions = {
  channelId: string;
};

const uploadToStorage = (
  file: File,
  uploadUrl: string,
  onProgress: (progress: number) => void,
): Promise<void> =>
  new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable) {
        const progress = Math.round((e.loaded / e.total) * 100);
        onProgress(progress);
      }
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(`アップロードに失敗しました (HTTP ${xhr.status})`));
      }
    });

    xhr.addEventListener("error", () => {
      reject(new Error("ネットワークエラーが発生しました"));
    });

    xhr.addEventListener("abort", () => {
      reject(new Error("アップロードがキャンセルされました"));
    });

    xhr.open("PUT", uploadUrl);
    xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");
    xhr.send(file);
  });

export const useFileUpload = () => {
  const [pendingAttachments, setPendingAttachments] = useState<PendingAttachment[]>([]);
  const presignMutation = usePresignUpload();

  const uploadFile = useCallback(
    async (file: File, options: UploadOptions): Promise<string | null> => {
      // バリデーション
      const validation = validateFile(file);
      if (!validation.valid) {
        setPendingAttachments((prev) => [
          ...prev,
          {
            file,
            state: { error: validation.error, status: "error" },
          },
        ]);
        return null;
      }

      // ペンディング状態を追加
      const pendingIndex = pendingAttachments.length;
      setPendingAttachments((prev) => [
        ...prev,
        {
          file,
          state: { status: "presigning" },
        },
      ]);

      try {
        // 表示時にレイアウトを予約できるよう、寸法と再生時間を送る
        const presignData = await presignMutation.mutateAsync({
          ...(await measureMedia(file)),
          channelId: options.channelId,
          contentType: file.type || "application/octet-stream",
          fileName: file.name,
          sizeBytes: BigInt(file.size),
        });

        // アップロード中に変更
        setPendingAttachments((prev) => {
          const next = [...prev];
          if (next[pendingIndex]) {
            next[pendingIndex] = {
              file: next[pendingIndex].file,
              state: { progress: 0, status: "uploading" },
            };
          }
          return next;
        });

        // ストレージへ直接アップロード
        await uploadToStorage(file, presignData.uploadUrl, (progress) => {
          setPendingAttachments((prev) => {
            const next = [...prev];
            if (next[pendingIndex]) {
              next[pendingIndex] = {
                file: next[pendingIndex].file,
                state: { progress, status: "uploading" },
              };
            }
            return next;
          });
        });

        // 完了状態に変更
        setPendingAttachments((prev) => {
          const next = [...prev];
          if (next[pendingIndex]) {
            next[pendingIndex] = {
              file: next[pendingIndex].file,
              state: { attachmentId: presignData.attachmentId, status: "completed" },
            };
          }
          return next;
        });

        return presignData.attachmentId;
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : "アップロードに失敗しました";

        setPendingAttachments((prev) => {
          const next = [...prev];
          if (next[pendingIndex]) {
            next[pendingIndex] = {
              file: next[pendingIndex].file,
              state: { error: errorMessage, status: "error" },
            };
          }
          return next;
        });

        return null;
      }
    },
    [pendingAttachments.length, presignMutation],
  );

  const removeAttachment = useCallback((index: number) => {
    setPendingAttachments((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const clearAttachments = useCallback(() => {
    setPendingAttachments([]);
  }, []);

  const getCompletedAttachmentIds = useCallback(
    (): string[] =>
      pendingAttachments
        .filter(
          (a): a is PendingAttachment & { state: { status: "completed"; attachmentId: string } } =>
            a.state.status === "completed",
        )
        .map((a) => a.state.attachmentId),
    [pendingAttachments],
  );

  return {
    clearAttachments,
    getCompletedAttachmentIds,
    isUploading: pendingAttachments.some(
      (a) => a.state.status === "uploading" || a.state.status === "presigning",
    ),
    pendingAttachments,
    removeAttachment,
    uploadFile,
  };
};
