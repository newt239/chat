import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { putToStorage } from "#/lib/upload";

type Listener = () => void;
type SentRequest = { method: string; url: string; contentType: string | null };

/** 送信したら status を返して load を、status が 0 なら error を発火する XMLHttpRequest に差し替える */
const stubXhr = (status: number) => {
  const requests: SentRequest[] = [];
  vi.stubGlobal(
    "XMLHttpRequest",
    class {
      public status = 0;
      public readonly upload = { addEventListener: () => {} };
      private readonly listeners = new Map<string, Listener>();
      private request: SentRequest = { contentType: null, method: "", url: "" };
      public addEventListener(type: string, listener: Listener) {
        this.listeners.set(type, listener);
      }
      public open(method: string, url: string) {
        this.request = { ...this.request, method, url };
      }
      public setRequestHeader(_name: string, value: string) {
        this.request.contentType = value;
      }
      public send() {
        requests.push(this.request);
        this.status = status;
        this.listeners.get(status === 0 ? "error" : "load")?.();
      }
    },
  );
  return requests;
};

describe("putToStorage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("Content-Type を付けて PUT する", async () => {
    const requests = stubXhr(200);
    await putToStorage(new Blob(["x"], { type: "image/png" }), "https://s3/put", () => {});
    expect(requests).toEqual([{ contentType: "image/png", method: "PUT", url: "https://s3/put" }]);
  });

  test("失敗したら画面に出せる文言で reject する", async () => {
    stubXhr(500);
    await expect(putToStorage(new Blob(["x"]), "https://s3/put", () => {})).rejects.toThrow(
      "アップロードに失敗しました（HTTP 500）",
    );
    stubXhr(0);
    await expect(putToStorage(new Blob(["x"]), "https://s3/put", () => {})).rejects.toThrow(
      "ネットワークエラーが発生しました",
    );
  });
});
