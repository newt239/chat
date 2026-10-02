import { useState } from "react";

import { useTranslation } from "react-i18next";

import { usePresignUpload } from "../api/client";
import { measureMedia } from "../utils/measureMedia";
import { formatFileSize, validateFile } from "../utils/validator";

import type { PendingAttachment } from "../api/types";

type UploadOptions = {
  channelId: string;
  // 録音のように、ファイルから再生時間を読めないときに計測済みの値を渡す
  durationSeconds: number | undefined;
};

// 進捗の通知と、失敗したときの文言（辞書から取ったもの）
type UploadHandlers = {
  onProgress: (progress: number) => void;
  http: (status: number) => string;
  network: string;
  aborted: string;
};

const uploadToStorage = (file: Blob, uploadUrl: string, handlers: UploadHandlers) =>
  new Promise<void>((resolve, reject) => {
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

  // 並行して上げても互いの行を上書きしないよう、添付ごとの id で更新する
  const setState = (id: string, state: PendingAttachment["state"]) => {
    setPendingAttachments((prev) =>
      prev.map((attachment) => (attachment.id === id ? { ...attachment, state } : attachment)),
    );
  };

  const uploadFile = async (file: File, options: UploadOptions) => {
    const id = crypto.randomUUID();
    const invalidReason = validateFile(file);
    if (invalidReason !== null) {
      setPendingAttachments((prev) => [
        ...prev,
        {
          file,
          id,
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

    setPendingAttachments((prev) => [...prev, { file, id, state: { status: "presigning" } }]);

    try {
      // 表示時にレイアウトを予約できるよう、寸法と再生時間を送る。動画はサムネイルも一緒に上げる
      const { thumbnail, ...measured } = await measureMedia(file);
      const presignData = await presignMutation.mutateAsync({
        ...measured,
        channelId: options.channelId,
        contentType: file.type || "application/octet-stream",
        durationSeconds: options.durationSeconds ?? measured.durationSeconds,
        fileName: file.name,
        sizeBytes: BigInt(file.size),
        thumbnail: thumbnail && {
          contentType: thumbnail.blob.type,
          height: thumbnail.height,
          sizeBytes: BigInt(thumbnail.blob.size),
          width: thumbnail.width,
        },
      });
      setState(id, { progress: 0, status: "uploading" });

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

      await uploadToStorage(file, presignData.uploadUrl, {
        ...errors,
        onProgress: (progress) => {
          setState(id, { progress, status: "uploading" });
        },
      });
      await thumbnailUpload;

      setState(id, { attachmentId: presignData.attachmentId, status: "completed" });
      return presignData.attachmentId;
    } catch (error) {
      setState(id, {
        error: error instanceof Error ? error.message : t("attachment.errors.unknown"),
        status: "error",
      });
      return null;
    }
  };

  const removeAttachment = (id: string) => {
    setPendingAttachments((prev) => prev.filter((attachment) => attachment.id !== id));
  };

  const clearAttachments = () => {
    setPendingAttachments([]);
  };

  const getCompletedAttachmentIds = () =>
    pendingAttachments.flatMap(({ state }) =>
      state.status === "completed" ? [state.attachmentId] : [],
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
