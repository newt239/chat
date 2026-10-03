import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { toast } from "#/components/ui/ToastRegion/toast";
import { toastError } from "#/lib/toastError";

vi.mock("#/components/ui/ToastRegion/toast", () => ({ toast: vi.fn() }));

describe("toastError", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("サーバーのメッセージを出す", () => {
    toastError(new Error("権限がありません"));
    expect(toast).toHaveBeenCalledWith("権限がありません", { tone: "danger" });
  });

  test("メッセージがなければ共通の文言を出す", () => {
    toastError(new Error(""));
    expect(toast).toHaveBeenCalledWith("操作できませんでした", { tone: "danger" });
  });
});
