import { useCallback, useState } from "react";

import { apiClient as api } from "../../../lib/api/client";

import type { LinkPreview, OGPData } from "../types";

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
        throw new Error(response.error.error || "Failed to fetch OGP");
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
      console.error("Failed to fetch OGP:", _error);
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
            error: ogpData ? undefined : "Failed to fetch preview",
          })
        );
      } catch (_error) {
        setPreviews((prev) =>
          new Map(prev).set(url, {
            url,
            ogpData: {},
            isLoading: false,
            error: "Failed to fetch preview",
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
