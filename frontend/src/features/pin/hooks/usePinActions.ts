import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import { api } from "@/lib/api/client";
import { store } from "@/providers/store";
import { addChannelPinsDeltaAtom } from "@/providers/store/ui";
import { currentChannelIdAtom } from "@/providers/store/workspace";

export function usePinActions(channelIdParam?: string | null) {
  const queryClient = useQueryClient();
  const addPinsDelta = useSetAtom(addChannelPinsDeltaAtom);
  const currentChannelId = channelIdParam ?? store.get(currentChannelIdAtom);

  const pin = useMutation({
    mutationFn: async ({ messageId }: { messageId: string }) => {
      if (!currentChannelId) throw new Error("チャンネルが選択されていません");
      // @ts-expect-error OpenAPI スキーマに pins が未反映
      const { data, error } = await api.POST("/api/channels/{channelId}/pins", {
        params: { path: { channelId: currentChannelId } },
        body: { messageId },
      });
      // @ts-expect-error error 型はスキーマ反映後に解消
      if (error) throw new Error(error.error);
      return data;
    },
    onSuccess: () => {
      if (!currentChannelId) return;
      queryClient.invalidateQueries({ queryKey: ["channels", currentChannelId, "pins"] });
      addPinsDelta({ channelId: currentChannelId, delta: 1 });
    },
  });

  const unpin = useMutation({
    mutationFn: async ({ messageId }: { messageId: string }) => {
      if (!currentChannelId) throw new Error("チャンネルが選択されていません");
      // @ts-expect-error OpenAPI スキーマに pins が未反映
      const { error } = await api.DELETE("/api/channels/{channelId}/pins/{messageId}", {
        params: { path: { channelId: currentChannelId, messageId } },
      });
      if (error) throw new Error(error.error);
      return { ok: true } as const;
    },
    onSuccess: () => {
      if (!currentChannelId) return;
      queryClient.invalidateQueries({ queryKey: ["channels", currentChannelId, "pins"] });
      addPinsDelta({ channelId: currentChannelId, delta: -1 });
    },
  });

  return { pin, unpin };
}
