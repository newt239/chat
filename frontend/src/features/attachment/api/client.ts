import { useMutation, useQuery } from "@tanstack/react-query";

import type { PresignRequest } from "./types";

import { api } from "@/lib/api/client";

export const usePresignUpload = () => {
  return useMutation({
    mutationFn: async (params: PresignRequest & { channelId: string }) => {
      const { data, error } = await api.POST("/api/attachments/presign", {
        body: params,
      });

      if (error) {
        throw new Error(error.error || "プリサイン URL の取得に失敗しました");
      }

      return data;
    },
  });
};

export const useAttachmentMetadata = (attachmentId: string) => {
  return useQuery({
    queryKey: ["attachment", attachmentId],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/attachments/{id}", {
        params: {
          path: { id: attachmentId },
        },
      });

      if (error) {
        throw new Error(error.error || "添付ファイル情報の取得に失敗しました");
      }

      return data;
    },
    enabled: Boolean(attachmentId),
  });
};

export const useDownloadUrl = () => {
  return useMutation({
    mutationFn: async (attachmentId: string) => {
      const { data, error } = await api.GET("/api/attachments/{id}/download", {
        params: {
          path: { id: attachmentId },
        },
      });

      if (error) {
        throw new Error(error.error || "ダウンロード URL の取得に失敗しました");
      }

      return data;
    },
  });
};
