import { useState, useCallback } from "react";

import { usePresignUpload } from "../api/client";
import { validateFile } from "../utils/validator";

import type { PendingAttachment } from "../api/types";

type UploadOptions = {
  channelId: string;
};

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
            state: { status: "error", error: validation.error },
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
        // プリサイン URL を取得
        const presignData = await presignMutation.mutateAsync({
          fileName: file.name,
          contentType: file.type || "application/octet-stream",
          sizeBytes: file.size,
          channelId: options.channelId,
        });

        // アップロード中に変更
        setPendingAttachments((prev) => {
          const next = [...prev];
          if (next[pendingIndex]) {
            next[pendingIndex] = {
              file: next[pendingIndex].file,
              state: { status: "uploading", progress: 0 },
            };
          }
          return next;
        });

        // Wasabi へ直接アップロード
        await uploadToWasabi(file, presignData.uploadUrl, (progress) => {
          setPendingAttachments((prev) => {
            const next = [...prev];
            if (next[pendingIndex]) {
              next[pendingIndex] = {
                file: next[pendingIndex].file,
                state: { status: "uploading", progress },
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
              state: { status: "completed", attachmentId: presignData.attachmentId },
            };
          }
          return next;
        });

        return presignData.attachmentId;
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : "アップロードに失敗しました";

        setPendingAttachments((prev) => {
          const next = [...prev];
          if (next[pendingIndex]) {
            next[pendingIndex] = {
              file: next[pendingIndex].file,
              state: { status: "error", error: errorMessage },
            };
          }
          return next;
        });

        return null;
      }
    },
    [pendingAttachments.length, presignMutation]
  );

  const removeAttachment = useCallback((index: number) => {
    setPendingAttachments((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const clearAttachments = useCallback(() => {
    setPendingAttachments([]);
  }, []);

  const getCompletedAttachmentIds = useCallback((): string[] => {
    return pendingAttachments
      .filter((a): a is PendingAttachment & { state: { status: "completed"; attachmentId: string } } =>
        a.state.status === "completed"
      )
      .map((a) => a.state.attachmentId);
  }, [pendingAttachments]);

  return {
    pendingAttachments,
    uploadFile,
    removeAttachment,
    clearAttachments,
    getCompletedAttachmentIds,
    isUploading: pendingAttachments.some(
      (a) => a.state.status === "uploading" || a.state.status === "presigning"
    ),
  };
};

const uploadToWasabi = async (
  file: File,
  uploadUrl: string,
  onProgress: (progress: number) => void
): Promise<void> => {
  return new Promise((resolve, reject) => {
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
};
