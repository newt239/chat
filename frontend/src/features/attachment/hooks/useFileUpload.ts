import { useState, useCallback } from "react";

import { useTranslation } from "react-i18next";

import { usePresignUpload } from "../api/client";
import { measureMedia } from "../utils/measureMedia";
import { formatFileSize, validateFile } from "../utils/validator";

import type { PendingAttachment } from "../api/types";

type UploadOptions = {
  channelId: string;
};

// 進捗の通知と、失敗したときの文言（辞書から取ったもの）
type UploadHandlers = {
  onProgress: (progress: number) => void;
  http: (status: number) => string;
  network: string;
  aborted: string;
};

const uploadToStorage = (file: Blob, uploadUrl: string, handlers: UploadHandlers): Promise<void> =>
  new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable) {
        const progress = Math.round((e.loaded / e.total) * 100);
        handlers.onProgress(progress);
      }
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(handlers.http(xhr.status)));
      }
    });

    xhr.addEventListener("error", () => {
      reject(new Error(handlers.network));
    });

    xhr.addEventListener("abort", () => {
      reject(new Error(handlers.aborted));
    });

    xhr.open("PUT", uploadUrl);
    xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");
    xhr.send(file);
  });

export const useFileUpload = () => {
  const [pendingAttachments, setPendingAttachments] = useState<PendingAttachment[]>([]);
  const presignMutation = usePresignUpload();
  const { t } = useTranslation();

  const uploadFile = useCallback(
    async (file: File, options: UploadOptions): Promise<string | null> => {
      // バリデーション
      const invalidReason = validateFile(file);
      if (invalidReason !== null) {
        setPendingAttachments((prev) => [
          ...prev,
          {
            file,
            state: {
              error:
                invalidReason === "empty"
                  ? t("attachment.errors.empty")
                  : t("attachment.errors.tooLarge", { size: formatFileSize(file.size) }),
              status: "error",
            },
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
        // 表示時にレイアウトを予約できるよう、寸法と再生時間を送る。動画はサムネイルも一緒に上げる
        const { thumbnail, ...media } = await measureMedia(file);
        const presignData = await presignMutation.mutateAsync({
          ...media,
          channelId: options.channelId,
          contentType: file.type || "application/octet-stream",
          fileName: file.name,
          sizeBytes: BigInt(file.size),
          thumbnail: thumbnail && {
            contentType: thumbnail.blob.type,
            height: thumbnail.height,
            sizeBytes: BigInt(thumbnail.blob.size),
            width: thumbnail.width,
          },
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

        const errors = {
          aborted: t("attachment.errors.aborted"),
          http: (status: number) => t("attachment.errors.http", { status }),
          network: t("attachment.errors.network"),
        };
        // サムネイルは表示を補うだけなので、失敗しても本体のアップロードは続ける
        const { thumbnailUploadUrl } = presignData;
        const thumbnailUpload =
          thumbnail && thumbnailUploadUrl !== undefined
            ? uploadToStorage(thumbnail.blob, thumbnailUploadUrl, {
                ...errors,
                onProgress: () => {},
              }).catch(() => {})
            : undefined;

        // ストレージへ直接アップロード
        await uploadToStorage(file, presignData.uploadUrl, {
          ...errors,
          onProgress: (progress) => {
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
          },
        });

        await thumbnailUpload;

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
        const errorMessage =
          error instanceof Error ? error.message : t("attachment.errors.unknown");

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
    [pendingAttachments.length, presignMutation, t],
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
