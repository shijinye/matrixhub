import { Button } from '@mantine/core'
import { IconUserPlus } from '@tabler/icons-react'
import {
  getRouteApi,
  useRouter,
  useRouterState,
} from '@tanstack/react-router'
import { useCallback } from 'react'
import { useTranslation } from 'react-i18next'

import { useRouteListState } from '@/shared/hooks/useRouteListState'

import { UsersTable } from '../components/UsersTable'

import type { User } from '@matrixhub/api-ts/v1alpha1/user.pb'

const usersRouteApi = getRouteApi('/(auth)/admin/users')

export function UsersPage() {
  const { t } = useTranslation()
  const router = useRouter()
  const navigate = usersRouteApi.useNavigate()
  const search = usersRouteApi.useSearch()
  const {
    users,
    pagination,
  } = usersRouteApi.useLoaderData()
  const loading = useRouterState({
    select: state => state.isLoading,
  })

  const refreshUsers = useCallback(() => {
    void router.invalidate({
      filter: match => match.routeId === '/(auth)/admin/users',
      sync: true,
    })
  }, [router])

  const {
    rowSelection,
    setRowSelection,
    selectedCount,
    handleSearchChange,
    handleRefresh,
    handlePageChange,
  } = useRouteListState({
    search,
    navigate,
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
      page={search.page ?? 1}
      searchValue={search.query ?? ''}
      onSearchChange={handleSearchChange}
      onRefresh={handleRefresh}
      onDelete={handleDelete}
      onBatchDelete={handleBatchDelete}
      rowSelection={rowSelection}
      onRowSelectionChange={setRowSelection}
      onPageChange={handlePageChange}
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
