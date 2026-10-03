const MAX_WIDTH = 400;
const MAX_HEIGHT = 300;
const MIN_WIDTH = 140;
const MIN_HEIGHT = 90;
// 寸法が分からない画像は 4:3 の枠に収める
const FALLBACK = { height: 300, width: 400 };

type ImageBox = {
  width: number;
  height: number;
  // 枠に収めるために切り取った向き
  crop: "tall" | "wide" | null;
};

// 1 枚だけの画像の表示枠。縮小しても最小サイズを下回る極端な比率のときは切り取る
export const imageBox = (width: number | undefined, height: number | undefined): ImageBox => {
  const source = width && height ? { height, width } : FALLBACK;
  const scale = Math.min(MAX_WIDTH / source.width, MAX_HEIGHT / source.height, 1);
  const scaledWidth = source.width * scale;
  const scaledHeight = source.height * scale;
  const isCropped = scaledWidth < MIN_WIDTH || scaledHeight < MIN_HEIGHT;
  return {
    crop: isCropped ? (source.height > source.width ? "tall" : "wide") : null,
    height: Math.round(Math.max(scaledHeight, MIN_HEIGHT)),
    width: Math.round(Math.max(scaledWidth, MIN_WIDTH)),
  };
};
