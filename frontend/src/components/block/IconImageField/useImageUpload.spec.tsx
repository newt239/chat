import type { ReactNode } from "react";

import { createRouterTransport } from "@connectrpc/connect";
import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { ImagePurpose, ImageService } from "#/gen/chat/v1/image_service_pb";
import { putToStorage } from "#/lib/upload";

import { useImageUpload } from "./useImageUpload";

import type { PresignImageUploadRequest } from "#/gen/chat/v1/image_service_pb";

vi.mock("#/lib/upload", () => ({ putToStorage: vi.fn(() => Promise.resolve()) }));

describe("useImageUpload", () => {
  test("発行された URL に画像を PUT して配信用の URL を返す", async () => {
    const presigned = vi.fn<(req: PresignImageUploadRequest) => void>();
    const transport = createRouterTransport((router) => {
      router.rpc(ImageService.method.presignImageUpload, (req) => {
        presigned(req);
        return { imageUrl: "https://api.example.com/images/a", uploadUrl: "https://s3/put" };
      });
    });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <TransportProvider transport={transport}>
        <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
      </TransportProvider>
    );
    const { result } = renderHook(() => useImageUpload(ImagePurpose.WORKSPACE_ICON, "ws1"), {
      wrapper,
    });

    const image = new Blob(["x"], { type: "image/webp" });
    let url = "";
    await act(async () => {
      url = await result.current.mutateAsync(image);
    });

    expect(url).toBe("https://api.example.com/images/a");
    expect(presigned).toHaveBeenCalledWith(
      expect.objectContaining({
        contentType: "image/webp",
        purpose: ImagePurpose.WORKSPACE_ICON,
        sizeBytes: 1n,
        workspaceId: "ws1",
      }),
    );
    expect(putToStorage).toHaveBeenCalledWith(image, "https://s3/put", expect.any(Function));
  });
});
