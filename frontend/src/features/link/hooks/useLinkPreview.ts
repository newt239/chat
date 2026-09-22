import { useCallback, useState } from "react";

import { api } from "#/lib/api/client";

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
        throw new Error(response.error.error || "OGPの取得に失敗しました");
      }

      return {
        cardType: response.data.ogpData.cardType || undefined,
        description: response.data.ogpData.description || undefined,
        imageUrl: response.data.ogpData.imageUrl || undefined,
        siteName: response.data.ogpData.siteName || undefined,
        title: response.data.ogpData.title || undefined,
      };
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
          isLoading: true,
          ogpData: {},
          url,
        });
      });

      try {
        const ogpData = await fetchOGP(url);

        setPreviews((prev) =>
          new Map(prev).set(url, {
            error: ogpData ? undefined : "プレビューの取得に失敗しました",
            isLoading: false,
            ogpData: ogpData ?? {},
            url,
          }),
        );
      } catch (error) {
        console.error("プレビューの取得に失敗しました:", error);
        setPreviews((prev) =>
          new Map(prev).set(url, {
            error: "プレビューの取得に失敗しました",
            isLoading: false,
            ogpData: {},
            url,
          }),
        );
      }
    },
    [fetchOGP],
  );

  const removePreview = useCallback((url: string) => {
    setPreviews((prev) => {
      const newMap = new Map(prev);
      newMap.delete(url);
      return newMap;
    });
  }, []);

  const getPreview = useCallback(
    (url: string): LinkPreview | undefined => previews.get(url),
    [previews],
  );

  const clearPreviews = useCallback(() => {
    setPreviews(new Map());
  }, []);

  const previewsArray: LinkPreview[] = [...previews.values()];

  return {
    addPreview,
    clearPreviews,
    getPreview,
    previews: previewsArray,
    removePreview,
  };
};
