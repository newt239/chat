import { canvasToBlob } from "#/lib/canvasToBlob";

import type { Area } from "react-easy-crop";

const OUTPUT_SIZE = 256;

// 切り抜いた範囲を 256px 四方の WebP にする。WebP に書き出せないブラウザでは PNG になる
export const cropImage = async (src: string, area: Area) => {
  const image = new Image();
  image.src = src;
  await image.decode();
  const canvas = document.createElement("canvas");
  canvas.width = OUTPUT_SIZE;
  canvas.height = OUTPUT_SIZE;
  canvas
    .getContext("2d")
    ?.drawImage(image, area.x, area.y, area.width, area.height, 0, 0, OUTPUT_SIZE, OUTPUT_SIZE);
  return canvasToBlob(canvas, "image/webp", 0.9);
};
