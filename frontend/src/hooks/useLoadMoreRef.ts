// 一覧の末尾に置いた要素が見えたら次のページを読み込む（無限スクロール）。戻り値は callback ref
export const useLoadMoreRef =
  (onLoadMore: () => void, isEnabled: boolean) => (element: HTMLElement | null) => {
    if (element === null || !isEnabled) {
      return undefined;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          onLoadMore();
        }
      },
      { rootMargin: "200px" },
    );
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  };
