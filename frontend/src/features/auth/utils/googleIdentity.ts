const scriptUrl = "https://accounts.google.com/gsi/client";

let loading: Promise<void> | null = null;

/** Google Identity Services のスクリプトを 1 度だけ読み込む。失敗したら次の呼び出しで再試行する */
export const loadGoogleIdentity = () => {
  loading ??= new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.src = scriptUrl;
    script.async = true;
    script.addEventListener("load", () => {
      resolve();
    });
    script.addEventListener("error", () => {
      loading = null;
      script.remove();
      reject(new Error("Google Identity Services を読み込めませんでした"));
    });
    document.head.append(script);
  });
  return loading;
};
