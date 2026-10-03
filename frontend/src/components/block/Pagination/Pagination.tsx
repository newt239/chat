import { IconChevronLeft, IconChevronRight } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";

type PaginationProps = {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
};

export const Pagination = ({ page, totalPages, onChange }: PaginationProps) => {
  const { t } = useTranslation();
  if (totalPages <= 1) {
    return null;
  }
  return (
    <nav className="flex items-center justify-center gap-2 pb-4 text-caption text-muted">
      <IconButton
        label={t("ui.pagination.previous")}
        isDisabled={page <= 1}
        onPress={() => {
          onChange(page - 1);
        }}
      >
        <IconChevronLeft />
      </IconButton>
      <span className="tabular-nums">{t("ui.pagination.page", { page, total: totalPages })}</span>
      <IconButton
        label={t("ui.pagination.next")}
        isDisabled={page >= totalPages}
        onPress={() => {
          onChange(page + 1);
        }}
      >
        <IconChevronRight />
      </IconButton>
    </nav>
  );
};
