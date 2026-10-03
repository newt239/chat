import { atom } from "jotai";

// モバイルで最後に開いたボトムタブ。チャンネルなどを開いている間も下に残す
export type MobileTab = "home" | "dms" | "activity" | "me";
export const mobileTabAtom = atom<MobileTab>("home");
