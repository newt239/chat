const oklabToLinearSrgb = (lightness: number, a: number, b: number) => {
  const l = (lightness + 0.3963377774 * a + 0.2158037573 * b) ** 3;
  const m = (lightness - 0.1055613458 * a - 0.0638541728 * b) ** 3;
  const s = (lightness - 0.0894841775 * a - 1.291485548 * b) ** 3;
  return [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.707614701 * s,
  ];
};

const isInGamut = (channels: number[]) => channels.every((v) => v >= -0.0005 && v <= 1.0005);

const toHexChannel = (linear: number) => {
  const v = Math.min(1, Math.max(0, linear));
  const srgb = v <= 0.0031308 ? 12.92 * v : 1.055 * v ** (1 / 2.4) - 0.055;
  return Math.round(srgb * 255)
    .toString(16)
    .padStart(2, "0");
};

// sRGB に収まるまで彩度を下げて hex に変換する（色相と明度は保つ）
export const oklchToHex = (lightness: number, chroma: number, hue: number) => {
  const radian = (hue * Math.PI) / 180;
  let channels = oklabToLinearSrgb(lightness, 0, 0);
  for (let c = chroma; c >= 0; c -= 0.004) {
    channels = oklabToLinearSrgb(lightness, c * Math.cos(radian), c * Math.sin(radian));
    if (isInGamut(channels)) {
      break;
    }
  }
  return `#${channels.map(toHexChannel).join("").toUpperCase()}`;
};

const relativeLuminance = (hex: string) => {
  const [r = 0, g = 0, b = 0] = [1, 3, 5].map((i) => {
    const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};

export const contrastRatio = (a: string, b: string) => {
  const [high, low] = [relativeLuminance(a), relativeLuminance(b)].toSorted((x, y) => y - x);
  return ((high ?? 0) + 0.05) / ((low ?? 0) + 0.05);
};
