import { atom } from "jotai";

type Session = {
  accessToken: string;
  userId: string;
};

// アクセストークンはメモリにだけ置く。リロード後は Cookie のリフレッシュトークンで取り直す
export const sessionAtom = atom<Session | null>(null);

export const myUserIdAtom = atom((get) => get(sessionAtom)?.userId ?? null);
