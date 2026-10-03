import { useRef } from "react";

import { IconChevronLeft, IconChevronRight, IconSearch } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { SearchField } from "#/components/ui/SearchField/SearchField";
import { Select } from "#/components/ui/Select/Select";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { Tab } from "#/components/ui/Tab/Tab";
import { TabList } from "#/components/ui/TabList/TabList";
import { TabPanel } from "#/components/ui/TabPanel/TabPanel";
import { Tabs } from "#/components/ui/Tabs/Tabs";
import { useWorkspaceSearch } from "#/features/search/hooks/useWorkspaceSearch";
import { searchFilterValues, searchSortValues } from "#/features/search/schemas";

import { SearchFilterBar } from "./SearchFilterBar";
import { SearchModifierHelp } from "./SearchModifierHelp";
import { SearchResultList } from "./SearchResultList";

import type { SearchFilter } from "#/features/search/schemas";

const RESULTS_PER_PAGE = 20;

const searchRoute = getRouteApi("/app/$workspaceId/search");

const pageCount = (total: number, perPage: number) =>
  Math.max(1, Math.ceil(total / Math.max(1, perPage)));

export const SearchPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = searchRoute.useParams();
  const search = searchRoute.useSearch();
  const { q: query, filter, page, sort } = search;
  const navigate = searchRoute.useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);

  const { data, isFetching, error, isEnabled, unresolved, resolved } = useWorkspaceSearch(
    workspaceId,
    search,
    RESULTS_PER_PAGE,
  );

  // 入力欄は URL のクエリを初期値にした非制御の欄なので、DOM の値を書き換えて input イベントで知らせる
  const insertModifier = (modifier: string) => {
    const input = inputRef.current;
    if (input === null) {
      return;
    }
    const prefix = input.value.trimEnd();
    input.setRangeText(`${prefix ? " " : ""}${modifier}`, prefix.length, input.value.length, "end");
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.focus();
  };

  const countOf = (value: SearchFilter) =>
    data === undefined
      ? 0
      : value === "all"
        ? Object.values(data).reduce((sum, section) => sum + section.total, 0)
        : data[value].total;
  const totalPages =
    data === undefined
      ? 0
      : Math.max(
          ...Object.entries(data)
            .filter(([key]) => filter === "all" || key === filter)
            .map(([, section]) => pageCount(section.total, section.perPage)),
        );

  const goToPage = (next: number) => {
    void navigate({ search: (prev) => ({ ...prev, page: next }) });
  };

  const renderResults = () => {
    const { invalidDates } = resolved.query;
    if (invalidDates.length > 0) {
      return (
        <p role="alert" className="m-0 px-4.5 py-6 text-caption text-danger">
          {t("search.invalidDate", { tokens: invalidDates.join(", ") })}
        </p>
      );
    }
    if (unresolved.length > 0) {
      return (
        <p role="alert" className="m-0 px-4.5 py-6 text-caption text-danger">
          {t("search.unresolved", { names: unresolved.join(", ") })}
        </p>
      );
    }
    if (query.trim().length === 0) {
      return <p className="m-0 px-4.5 py-6 text-caption text-muted">{t("search.prompt")}</p>;
    }
    if (error) {
      return (
        <p role="alert" className="m-0 px-4.5 py-6 text-caption text-danger">
          {t("search.failed")}
        </p>
      );
    }
    if (!isEnabled || isFetching || data === undefined) {
      return (
        <div className="flex flex-col gap-3 px-4.5 py-4">
          {[0, 1, 2].map((index) => (
            <Skeleton key={index} className="h-16 w-full rounded-lg" />
          ))}
        </div>
      );
    }
    if (countOf(filter) === 0) {
      return (
        <div className="flex flex-col items-center gap-1 px-4.5 py-10 text-center">
          <b className="text-body-strong">{t("search.empty")}</b>
          <span className="text-caption text-muted">{t("search.emptyHint")}</span>
        </div>
      );
    }
    return (
      <>
        <div className="flex min-h-10 items-center justify-between gap-2.5 px-4.5 pt-2.5 text-xs text-muted">
          {t("search.count", { count: countOf(filter) })}
          {(filter === "all" || filter === "messages") && (
            <Select
              label={t("search.sort.label")}
              options={searchSortValues.map((value) => ({
                label: t(`search.sort.${value}`),
                value,
              }))}
              value={sort}
              onChange={(next) => {
                void navigate({ search: (prev) => ({ ...prev, page: 1, sort: next }) });
              }}
              className="flex-row items-center gap-2 [&_button]:h-7 [&_button]:w-28 [&_button]:text-label [&_button]:font-normal [&_label]:text-xs [&_label]:font-normal [&_label]:text-muted"
            />
          )}
        </div>
        <SearchResultList
          messages={data.messages.items}
          channels={data.channels.items}
          users={data.users.items}
          groups={data.groups.items}
          filter={filter}
          workspaceId={workspaceId}
        />
        {totalPages > 1 && (
          <nav className="flex items-center justify-center gap-2 pb-6 text-caption text-muted">
            <IconButton
              label={t("search.prev")}
              isDisabled={page <= 1}
              onPress={() => {
                goToPage(page - 1);
              }}
            >
              <IconChevronLeft />
            </IconButton>
            <span className="tabular-nums">{t("search.page", { page, total: totalPages })}</span>
            <IconButton
              label={t("search.next")}
              isDisabled={page >= totalPages}
              onPress={() => {
                goToPage(page + 1);
              }}
            >
              <IconChevronRight />
            </IconButton>
          </nav>
        )}
      </>
    );
  };

  return (
    <section className="flex h-full min-h-0 flex-col bg-surface font-sans text-text">
      <PageHeader icon={<IconSearch aria-hidden />} title={t("search.title")} />
      <Form
        role="search"
        className="flex shrink-0 flex-col gap-2 px-4.5 pt-3 pb-2.5"
        onSubmit={(event) => {
          event.preventDefault();
          const q = inputRef.current?.value.trim() ?? "";
          void navigate({ search: (prev) => ({ ...prev, page: 1, q }) });
        }}
      >
        <SearchField
          key={query}
          defaultValue={query}
          label={t("search.input")}
          placeholder={t("search.placeholder")}
          inputRef={inputRef}
          className="h-10 rounded-lg pr-1.5 pl-3 text-title font-normal max-md:h-10"
        />
        <SearchModifierHelp onInsert={insertModifier} />
      </Form>
      <SearchFilterBar resolved={resolved} />
      <Tabs
        selectedKey={filter}
        onSelectionChange={(key) => {
          const next = searchFilterValues.find((value) => value === key);
          if (next !== undefined) {
            void navigate({ search: (prev) => ({ ...prev, filter: next, page: 1 }) });
          }
        }}
        className="min-h-0 flex-1"
      >
        <TabList aria-label={t("search.tabs")}>
          {searchFilterValues.map((value) => (
            <Tab key={value} id={value}>
              {t(`search.sections.${value}`)}
              {data !== undefined && (filter === "all" || value === filter) && (
                <small className="font-mono text-caption font-normal text-subtle">
                  {countOf(value)}
                </small>
              )}
            </Tab>
          ))}
        </TabList>
        {searchFilterValues.map((value) => (
          <TabPanel key={value} id={value} className="overflow-y-auto">
            {renderResults()}
          </TabPanel>
        ))}
      </Tabs>
    </section>
  );
};
