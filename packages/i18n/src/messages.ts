import type { ja } from "./locales/ja";

// 日本語の複数形は _other だけなので、他言語には対になる _one も求める
type Widen<T> = { [K in keyof T]: T[K] extends string ? string : Widen<T[K]> } & {
  [K in keyof T as K extends `${infer Base}_other` ? `${Base}_one` : never]: string;
};

// 日本語辞書と同じキーを持つことを他言語の辞書に強制する
export type Messages = Widen<typeof ja>;
