import { useCallback, useState } from "react";

import type { LinkPreview, OGPData } from "../types";

import { api } from "@/lib/api/client";

type UseLinkPreviewReturn = {
  previews: LinkPreview[];
  addPreview: (url: string) => Promise<void>;
  removePreview: (url: string) => void;
  getPreview: (url: string) => LinkPreview | undefined;
  clearPreviews: () => void;
};

export const useLinkPreview = (): UseLinkPreviewReturn => {
  const [previews, setPreviews] = useState<Map<string, LinkPreview>>(new Map());

  const fetchOGP = useCallback(async (url: string): Promise<OGPData | null> => {
    try {
      const response = await api.POST("/api/links/fetch-ogp", {
        body: { url },
      });

      if (response.error) {
        throw new Error(response.error.error || "OGPの取得に失敗しました");
      }

      return response.data?.ogpData
        ? {
            title: response.data.ogpData.title || undefined,
            description: response.data.ogpData.description || undefined,
            imageUrl: response.data.ogpData.imageUrl || undefined,
            siteName: response.data.ogpData.siteName || undefined,
            cardType: response.data.ogpData.cardType || undefined,
          }
        : null;
    } catch (_error) {
      console.error("OGPの取得に失敗しました:", _error);
      return null;
    }
  }, []);

  const addPreview = useCallback(
    async (url: string) => {
      // ローディング状態でプレビューを追加
      setPreviews((prev) => {
        // 既にプレビューが存在する場合は何もしない
        if (prev.has(url)) {
          return prev;
        }
        return new Map(prev).set(url, {
          url,
          ogpData: {},
          isLoading: true,
        });
      });

      try {
        const ogpData = await fetchOGP(url);

        setPreviews((prev) =>
          new Map(prev).set(url, {
            url,
            ogpData: ogpData || {},
            isLoading: false,
            error: ogpData ? undefined : "プレビューの取得に失敗しました",
          })
        );
      } catch (error) {
        console.error("プレビューの取得に失敗しました:", error);
        setPreviews((prev) =>
          new Map(prev).set(url, {
            url,
            ogpData: {},
            isLoading: false,
            error: "プレビューの取得に失敗しました",
          })
        );
      }
    },
    [fetchOGP]
  );

  const removePreview = useCallback((url: string) => {
    setPreviews((prev) => {
      const newMap = new Map(prev);
      newMap.delete(url);
      return newMap;
    });
  }, []);

  const getPreview = useCallback(
    (url: string): LinkPreview | undefined => {
      return previews.get(url);
    },
    [previews]
  );

  const clearPreviews = useCallback(() => {
    setPreviews(new Map());
  }, []);

  const previewsArray: LinkPreview[] = Array.from(previews.values());

  return {
    previews: previewsArray,
    addPreview,
    removePreview,
    getPreview,
    clearPreviews,
  };
};
