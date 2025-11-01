import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

import type { components } from "@/lib/api/schema";

import { storage } from "@/lib/storage";

type User = components["schemas"]["User"];

type AuthState = {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
};

const storageKey = "auth-storage";

const createEmptyAuthState = (): AuthState => ({
  user: null,
  accessToken: null,
  refreshToken: null,
});

const sanitizeAuthState = (state: Partial<AuthState>): AuthState => ({
  user: state.user ?? null,
  accessToken: state.accessToken ?? null,
  refreshToken: state.refreshToken ?? null,
});

const authStorageAtom = atomWithStorage<AuthState>(storageKey, createEmptyAuthState(), undefined, {
  getOnInit: true,
});

export const authAtom = atom(
  (get) => sanitizeAuthState(get(authStorageAtom)),
  (_get, set, update: AuthState) => {
    set(authStorageAtom, sanitizeAuthState(update));
  }
);

export const userAtom = atom<User | null>((get) => get(authAtom).user);
export const accessTokenAtom = atom<string | null>((get) => get(authAtom).accessToken);
export const refreshTokenAtom = atom<string | null>((get) => get(authAtom).refreshToken);
export const isAuthenticatedAtom = atom<boolean>((get) => {
  const state = get(authAtom);
  return Boolean(state.user && state.accessToken && state.refreshToken);
});

type AuthPayload = {
  user: User;
  accessToken: string;
  refreshToken: string;
};

export const setAuthAtom = atom(null, (_get, set, payload: AuthPayload) => {
  set(authAtom, sanitizeAuthState(payload));
});

export const clearAuthAtom = atom(null, (_get, set) => {
  set(authAtom, createEmptyAuthState());
});

export const initializeAuthAtom = atom(null, (get, set) => {
  const current = get(authAtom);

  const legacyAccessToken = storage.getItem("accessToken");
  const legacyRefreshToken = storage.getItem("refreshToken");

  if (!current.accessToken && !current.refreshToken && legacyAccessToken && legacyRefreshToken) {
    set(
      authAtom,
      sanitizeAuthState({
        user: current.user,
        accessToken: legacyAccessToken,
        refreshToken: legacyRefreshToken,
      })
    );
  }

  storage.removeItem("accessToken");
  storage.removeItem("refreshToken");
});
