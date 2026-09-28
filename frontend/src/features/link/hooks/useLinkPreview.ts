import { useCallback, useState } from "react";

import { useMutation } from "@connectrpc/connect-query";

import { LinkService } from "#/gen/chat/v1/link_service_pb";

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

  const { mutateAsync: fetchOgp } = useMutation(LinkService.method.fetchOgp);

  const fetchOGP = useCallback(
    async (url: string): Promise<OGPData | null> => {
      try {
        const { ogp } = await fetchOgp({ url });

        return {
          cardType: ogp?.cardType || undefined,
          description: ogp?.description || undefined,
          imageUrl: ogp?.imageUrl || undefined,
          siteName: ogp?.siteName || undefined,
          title: ogp?.title || undefined,
        };
      } catch (_error) {
        console.error("OGPの取得に失敗しました:", _error);
        return null;
      }
    },
    [fetchOgp],
  );

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
