import { useEffect } from "react";

import { useSetAtom } from "jotai";

import { initializeAuthAtom } from "@/providers/store/auth";

export const AuthInitializer = () => {
  const initializeAuth = useSetAtom(initializeAuthAtom);

  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  return null;
};
