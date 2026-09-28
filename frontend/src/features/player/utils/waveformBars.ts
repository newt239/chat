const BAR_COUNT = 44;
const MODULUS = 2_147_483_647;

// 文字列から 0.25〜1 の高さの列を決定的に作る（Park–Miller の線形合同法）
export const waveformBars = (seed: string) => {
  let state = 0;
  for (const char of seed) {
    state = (state * 31 + (char.codePointAt(0) ?? 0)) % MODULUS;
  }
  state = (state % (MODULUS - 1)) + 1;
  const random = () => {
    state = (state * 16_807) % MODULUS;
    return state / MODULUS;
  };
  return Array.from(
    { length: BAR_COUNT },
    (_, index) =>
      0.25 + 0.75 * Math.abs(Math.sin(index * 0.5 + random() * 2)) * (0.5 + random() * 0.5),
  );
};
