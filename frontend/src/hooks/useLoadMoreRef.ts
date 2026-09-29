import { useEffect, useEffectEvent, useRef } from "react";

// 一覧の末尾に置いた要素が見えたら次のページを読み込む（無限スクロール）
export const useLoadMoreRef = (onLoadMore: () => void, isEnabled: boolean) => {
  const ref = useRef<HTMLDivElement>(null);
  const handleLoadMore = useEffectEvent(onLoadMore);

  useEffect(() => {
    const element = ref.current;
    if (element === null || !isEnabled) {
      return undefined;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          handleLoadMore();
        }
      },
      { rootMargin: "200px" },
    );
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  }, [isEnabled]);

  return ref;
};
