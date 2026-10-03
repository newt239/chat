// 時間は ms
type TimingToken = {
  type: "timing";
  duration: number;
  easing: readonly [number, number, number, number];
};

type SpringToken = {
  type: "spring";
  stiffness: number;
  damping: number;
  mass: number;
};

export type MotionToken = TimingToken | SpringToken;

const standardEasing = [0.2, 0, 0, 1] as const;

export const motion = {
  // ツリーの開閉・タブ内容
  base: { duration: 140, easing: standardEasing, type: "timing" },
  // フェード・ツールチップ
  fast: { duration: 90, easing: standardEasing, type: "timing" },
  // モバイルの画面遷移
  push: { damping: 50, mass: 1, stiffness: 600, type: "spring" },
  // シート・モバイルのダイアログ
  sheet: { damping: 44, mass: 1, stiffness: 560, type: "spring" },
  // ポップオーバー・選択ピル・リアクション
  spring: { damping: 42, mass: 0.9, stiffness: 700, type: "spring" },
} as const satisfies Record<string, MotionToken>;

export type MotionTokenName = keyof typeof motion;
