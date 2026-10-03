import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { toast } from "#/components/ui/ToastRegion/toast";
import { copyWithToast } from "#/lib/clipboard";

vi.mock("#/components/ui/ToastRegion/toast", () => ({ toast: vi.fn() }));

const stubClipboard = (writeText: () => Promise<void>) => {
  Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText } });
};

describe("copyWithToast", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("書き込めたら成功の文言を通知する", async () => {
    stubClipboard(() => Promise.resolve());
    await copyWithToast("text", "コピーしました");
    expect(toast).toHaveBeenCalledWith("コピーしました", { tone: "success" });
  });

  test("書き込めなければ成功を通知せず、失敗を通知する", async () => {
    stubClipboard(() => Promise.reject(new Error("denied")));
    await copyWithToast("text", "コピーしました");
    expect(toast).toHaveBeenCalledOnce();
    expect(toast).toHaveBeenCalledWith("コピーできませんでした", { tone: "danger" });
  });
});
