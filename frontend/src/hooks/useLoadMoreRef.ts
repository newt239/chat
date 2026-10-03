type LoadMoreQuery = {
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => void;
};

// 一覧の末尾に置いた要素が見えたら次のページを読み込む（無限スクロール）。戻り値は callback ref
export const useLoadMoreRef =
  ({ hasNextPage, isFetchingNextPage, fetchNextPage }: LoadMoreQuery) =>
  (element: HTMLElement | null) => {
    if (element === null || !hasNextPage || isFetchingNextPage) {
      return undefined;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          fetchNextPage();
        }
      },
      { rootMargin: "200px" },
    );
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  };
