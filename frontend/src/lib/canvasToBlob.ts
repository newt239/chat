export const canvasToBlob = (
  canvas: HTMLCanvasElement,
  type: string,
  quality: number | undefined,
) =>
  new Promise<Blob>((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob === null) {
          reject(new Error("failed to encode image"));
        } else {
          resolve(blob);
        }
      },
      type,
      quality,
    );
  });
