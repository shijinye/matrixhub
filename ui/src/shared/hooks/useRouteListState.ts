import {
  startTransition,
  useState,
} from 'react'

import type { MRT_RowSelectionState } from 'mantine-react-table'

interface RouteListSearchState {
  page?: number
  query?: string
}

interface UseRouteListStateOptions<TSearch extends RouteListSearchState, TRecord> {
  search: TSearch
  navigate: (options: {
    replace?: boolean
    search: (prev: TSearch) => TSearch
  }) => unknown
  records: readonly TRecord[]
  getRecordId: (record: TRecord) => string
  refresh?: () => unknown
  defaultPage?: number
  normalizeQuery?: (value: string) => string | undefined
}

const DEFAULT_PAGE = 1

function defaultNormalizeQuery(value: string) {
  const nextQuery = value.trim()

  return nextQuery.length > 0 ? nextQuery : undefined
}

export function useRouteListState<TSearch extends RouteListSearchState, TRecord>({
  search,
  navigate,
  records,
  getRecordId,
  refresh,
  defaultPage = DEFAULT_PAGE,
  normalizeQuery = defaultNormalizeQuery,
}: UseRouteListStateOptions<TSearch, TRecord>) {
  const [rowSelection, setRowSelection] = useState<MRT_RowSelectionState>({})

  const currentRecordIds = new Set(records.map(record => getRecordId(record)))

  const clearRowSelection = () => {
    setRowSelection({})
  }

  const selectedRowIds = Object.keys(rowSelection).filter(
    rowId => !!rowSelection[rowId] && currentRecordIds.has(rowId),
  )

  const selectedCount = selectedRowIds.length
  const currentQuery = normalizeQuery(search.query ?? '')

  const onSearchChange = (value: string) => {
    const nextQuery = normalizeQuery(value)

    if (nextQuery === currentQuery) {
      return
    }

    clearRowSelection()
    void navigate({
      replace: true,
      search: prev => ({
        ...prev,
        page: defaultPage,
        query: nextQuery,
      }),
    })
  }

  const onRefresh = () => {
    clearRowSelection()
    void refresh?.()
  }

  const onPageChange = (page: number) => {
    if (page === (search.page ?? defaultPage)) {
      return
    }

    clearRowSelection()
    startTransition(() => {
      void navigate({
        search: prev => ({
          ...prev,
          page,
        }),
      })
    })
  }

  return {
    rowSelection,
    setRowSelection,
    clearRowSelection,
    selectedCount,
    selectedRowIds,
    onSearchChange,
    onRefresh,
    onPageChange,
  }
}
