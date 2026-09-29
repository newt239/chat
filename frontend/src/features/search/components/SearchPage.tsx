import { useRef, useState } from "react";

import { IconChevronLeft, IconChevronRight, IconSearch } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { Form, Input, SearchField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { IconButton } from "#/components/ui/IconButton/IconButton";
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

  const [inputValue, setInputValue] = useState(query);
  const [syncedQuery, setSyncedQuery] = useState(query);
  // URL のクエリが変わったら入力欄に反映する
  if (syncedQuery !== query) {
    setSyncedQuery(query);
    setInputValue(query);
  }

  const { data, isFetching, error, isEnabled, unresolved, resolved } = useWorkspaceSearch(
    workspaceId,
    search,
    RESULTS_PER_PAGE,
  );

  const insertModifier = (modifier: string) => {
    setInputValue((prev) => `${prev.trimEnd()}${prev.trim() ? " " : ""}${modifier}`);
    requestAnimationFrame(() => {
      const input = inputRef.current;
      input?.focus();
      input?.setSelectionRange(input.value.length, input.value.length);
    });
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
        <p role="alert" className="m-0 px-[18px] py-6 text-caption text-danger">
          {t("search.invalidDate", { tokens: invalidDates.join(", ") })}
        </p>
      );
    }
    if (unresolved.length > 0) {
      return (
        <p role="alert" className="m-0 px-[18px] py-6 text-caption text-danger">
          {t("search.unresolved", { names: unresolved.join(", ") })}
        </p>
      );
    }
    if (query.trim().length === 0) {
      return <p className="m-0 px-[18px] py-6 text-caption text-muted">{t("search.prompt")}</p>;
    }
    if (error) {
      return (
        <p role="alert" className="m-0 px-[18px] py-6 text-caption text-danger">
          {t("search.failed")}
        </p>
      );
    }
    if (!isEnabled || isFetching || data === undefined) {
      return (
        <div className="flex flex-col gap-3 px-[18px] py-4">
          {[0, 1, 2].map((index) => (
            <Skeleton key={index} className="h-16 w-full rounded-lg" />
          ))}
        </div>
      );
    }
    if (countOf(filter) === 0) {
      return (
        <div className="flex flex-col items-center gap-1 px-[18px] py-10 text-center">
          <b className="text-body-strong">{t("search.empty")}</b>
          <span className="text-caption text-muted">{t("search.emptyHint")}</span>
        </div>
      );
    }
    return (
      <>
        <div className="flex min-h-10 items-center justify-between gap-2.5 px-[18px] pt-2.5 text-xs text-muted">
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
              className="flex-row items-center gap-2 [&_button]:h-7 [&_button]:w-28 [&_button]:text-[12.5px] [&_label]:text-xs [&_label]:font-normal [&_label]:text-muted"
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
        className="flex shrink-0 flex-col gap-2 px-[18px] pt-3 pb-2.5"
        onSubmit={(event) => {
          event.preventDefault();
          void navigate({ search: (prev) => ({ ...prev, page: 1, q: inputValue.trim() }) });
        }}
      >
        <SearchField
          value={inputValue}
          onChange={setInputValue}
          aria-label={t("search.input")}
          className="flex h-10 items-center gap-2 rounded-[10px] border border-border-strong bg-surface pr-1.5 pl-3 text-muted data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft"
        >
          <IconSearch aria-hidden className="size-4 shrink-0" />
          <Input
            ref={inputRef}
            placeholder={t("search.placeholder")}
            className="h-full min-w-0 flex-1 border-0 bg-transparent font-sans text-[15px] text-text outline-none placeholder:text-subtle [&::-webkit-search-cancel-button]:hidden"
          />
        </SearchField>
        <SearchModifierHelp onInsert={insertModifier} />
      </Form>
      <SearchFilterBar />
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
                <small className="font-mono text-[11px] font-normal text-subtle">
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
