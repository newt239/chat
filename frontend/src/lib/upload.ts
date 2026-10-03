import { i18n } from "#/lib/i18n";

/** 署名付き URL へ直接 PUT する。進捗は 0〜100 で知らせ、失敗したら画面に出せる文言で reject する */
export const putToStorage = (
  blob: Blob,
  uploadUrl: string,
  onProgress: (progress: number) => void,
) =>
  new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    });
    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(i18n.t("common.upload.http", { status: xhr.status })));
      }
    });
    xhr.addEventListener("error", () => {
      reject(new Error(i18n.t("common.upload.network")));
    });
    xhr.addEventListener("abort", () => {
      reject(new Error(i18n.t("common.upload.aborted")));
    });
    xhr.open("PUT", uploadUrl);
    xhr.setRequestHeader("Content-Type", blob.type || "application/octet-stream");
    xhr.send(blob);
  });
