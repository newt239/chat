// 時間は ms。Web は Motion、Native は Reanimated（withTiming / withSpring）に同じ値を渡す
export type TimingToken = {
  type: "timing";
  duration: number;
  easing: readonly [number, number, number, number];
};

export type SpringToken = {
  type: "spring";
  stiffness: number;
  damping: number;
  mass: number;
};

export type MotionToken = TimingToken | SpringToken;

const standardEasing = [0.2, 0, 0, 1] as const;

export const motion = {
  // ツリーの開閉・タブ内容
  base: { duration: 200, easing: standardEasing, type: "timing" },
  // フェード・ツールチップ
  fast: { duration: 120, easing: standardEasing, type: "timing" },
  // モバイルの画面遷移
  push: { damping: 44, mass: 1, stiffness: 420, type: "spring" },
  // シート・モバイルのダイアログ
  sheet: { damping: 38, mass: 1, stiffness: 380, type: "spring" },
  // ポップオーバー・選択ピル・リアクション
  spring: { damping: 40, mass: 0.9, stiffness: 520, type: "spring" },
} as const satisfies Record<string, MotionToken>;

export type MotionTokenName = keyof typeof motion;
