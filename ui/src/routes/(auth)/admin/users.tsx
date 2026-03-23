import { Users } from '@matrixhub/api-ts/v1alpha1/user.pb'
import { IconUsers as AdminUsersIcon } from '@tabler/icons-react'
import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { AdminPageLayout } from '@/features/admin/components/admin-page-layout'
import { UsersPage } from '@/features/admin/users/pages/UsersPage'

const DEFAULT_USERS_PAGE = 1
const DEFAULT_USERS_PAGE_SIZE = 10

const usersSearchSchema = z.object({
  page: z.coerce.number().int().gte(1).optional().catch(DEFAULT_USERS_PAGE),
  query: z.string().trim().optional().catch(undefined),
})

export const Route = createFileRoute('/(auth)/admin/users')({
  validateSearch: usersSearchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    query: search.query,
  }),
  loader: async ({ deps }) => {
    // TODO: Need to handle errors or disaplay errorComponent
    const response = await Users.ListUsers({
      page: deps.page,
      pageSize: DEFAULT_USERS_PAGE_SIZE,
      search: deps.query,
    })

    return {
      users: response.users ?? [],
      pagination: response.pagination,
    }
  },
  component: RouteComponent,
})

export const Icon = AdminUsersIcon

function RouteComponent() {
  const { t } = useTranslation()

  return (
    <AdminPageLayout
      icon={AdminUsersIcon}
      title={t('admin.users')}
    >
      <UsersPage />
    </AdminPageLayout>
  )
}
