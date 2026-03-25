import { Button } from '@mantine/core'
import { IconUserPlus } from '@tabler/icons-react'
import {
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import {
  getRouteApi,
  useRouterState,
} from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useRouteListState } from '@/shared/hooks/useRouteListState'

import { UsersTable } from '../components/UsersTable'
import {
  adminUserKeys,
  usersQueryOptions,
} from '../users.query'
import { DEFAULT_USERS_PAGE } from '../users.schema'
import { getUserRowId } from '../users.utils'

import type { User } from '@matrixhub/api-ts/v1alpha1/user.pb'

const usersRouteApi = getRouteApi('/(auth)/admin/users')

export function UsersPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const navigate = usersRouteApi.useNavigate()
  const search = usersRouteApi.useSearch()
  const {
    data,
    isFetching,
  } = useSuspenseQuery(usersQueryOptions(search))
  const {
    users,
    pagination,
  } = data
  const routeLoading = useRouterState({
    select: state => state.isLoading,
  })
  const loading = routeLoading || isFetching

  const refreshUsers = () => queryClient.invalidateQueries({
    queryKey: adminUserKeys.lists(),
  })

  const {
    rowSelection,
    setRowSelection,
    selectedCount,
    onSearchChange,
    onRefresh,
    onPageChange,
  } = useRouteListState({
    search,
    navigate,
    records: users,
    getRecordId: getUserRowId,
    refresh: refreshUsers,
  })

  const handleCreate = () => {
    // TODO: open create user modal
  }

  const handleDelete = (_user: User) => {
    // TODO: Implement delete functionality
  }

  const handleBatchDelete = () => {
    if (selectedCount === 0) {
      return
    }

    // TODO: open batch delete user modal
  }

  return (
    <UsersTable
      records={users}
      pagination={pagination}
      loading={loading}
      page={search.page ?? DEFAULT_USERS_PAGE}
      searchValue={search.query ?? ''}
      onSearchChange={onSearchChange}
      onRefresh={onRefresh}
      onDelete={handleDelete}
      onBatchDelete={handleBatchDelete}
      rowSelection={rowSelection}
      onRowSelectionChange={setRowSelection}
      onPageChange={onPageChange}
      selectedCount={selectedCount}
      toolbarExtra={(
        <Button
          disabled
          onClick={handleCreate}
          leftSection={<IconUserPlus size={16} />}
        >
          {t('routes.admin.users.toolbar.create')}
        </Button>
      )}
    />
  )
}
