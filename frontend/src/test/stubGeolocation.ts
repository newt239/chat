import { vi } from "vite-plus/test";

type Coordinates = { latitude: number; longitude: number; accuracy: number };
type Success = (position: { coords: Coordinates }) => void;
type Failure = (error: { code: number; PERMISSION_DENIED: number }) => void;

// jsdom には Geolocation API がないため、respond で成功・失敗を返す実装を navigator に生やす
export const stubGeolocation = (respond: (success: Success, failure: Failure) => void) => {
  const getCurrentPosition = vi.fn(respond);
  Object.defineProperty(navigator, "geolocation", {
    configurable: true,
    value: { getCurrentPosition },
  });
  return getCurrentPosition;
};
