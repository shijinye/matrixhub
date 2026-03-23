import {
  Box,
  Button,
  Tabs,
} from '@mantine/core'
import { Models } from '@matrixhub/api-ts/v1alpha1/model.pb.ts'
import { IconDownload, IconCloudUpload } from '@tabler/icons-react'
import {
  Outlet, useMatchRoute, createFileRoute, linkOptions, Link,
} from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { projectRolesQueryOptions } from '@/features/auth/auth.query'
import { buildModelBadges, buildModelMetaItems } from '@/features/models/models.utils'
import { queryClient } from '@/queryClient'
import { ResourceDetailHeader } from '@/shared/components/ResourceDetailHeader'
import { notFoundError } from '@/utils/routerAccess'

import { Route as ModelSettingsRoute } from './settings'
import { Route as ModelTreeRoute } from './tree/$ref/$'

import { Route as ModelIndexRoute } from './index'

export const Route = createFileRoute(
  '/(auth)/(app)/projects_/$projectId/models/$modelId',
)({
  component: ModelDetailLayout,
  loader: async ({ params }) => {
    const projectRoles = await queryClient.ensureQueryData(projectRolesQueryOptions())
    let model: Awaited<ReturnType<typeof Models.GetModel>>

    // if the project is private and user has no access, will throw error
    try {
      model = await Models.GetModel({
        project: params.projectId,
        name: params.modelId,
      })
    } catch {
      throw notFoundError()
    }

    return {
      model,
      projectRoles: projectRoles,
    }
  },
})

function ModelDetailLayout() {
  const { t } = useTranslation()
  const {
    projectId, modelId,
  } = Route.useParams()

  const {
    model, projectRoles,
  } = Route.useLoaderData()
  const hasProjectRole = Boolean(projectRoles.projectRoles?.[projectId])

  const tabRoutes = linkOptions([
    {
      id: 'desc',
      label: t('model.detail.desc'),
      to: ModelIndexRoute.to,
      params: {
        projectId,
        modelId,
      },
    },
    {
      id: 'tree',
      label: t('model.detail.tree'),
      to: ModelTreeRoute.to,
      params: {
        projectId,
        modelId,
        // TODO: use real ref
        ref: 'main',
      },
    },
    ...(hasProjectRole
      ? [{
          id: 'settings',
          label: t('model.detail.setting'),
          to: ModelSettingsRoute.to,
          params: {
            projectId,
            modelId,
          },
        }]
      : []),
  ])

  const matchRoute = useMatchRoute()
  const activeTab = tabRoutes.find(tab => matchRoute({ to: tab.to }))?.id || tabRoutes[0].id

  return (
    <Box pt={20} pb={32}>
      <Box mb={24}>
        <ResourceDetailHeader
          projectId={projectId}
          name={modelId}
          badges={buildModelBadges(model)}
          metaItems={buildModelMetaItems(model, projectId)}
          actions={(
            <>
              <Button size="xs" color="cyan" variant="light" leftSection={<IconCloudUpload size={16} />}>{t('model.upload')}</Button>
              <Button size="xs" color="cyan" variant="light" leftSection={<IconDownload size={16} />}>{t('model.download')}</Button>
            </>
          )}
        />
      </Box>
      <Tabs value={activeTab}>
        <Tabs.List>
          {
            tabRoutes.map(({
              id, label, ...linkProps
            }) => (
              <Tabs.Tab
                key={id}
                value={id}
                component={Link}
                {...linkProps}
              >
                {label}
              </Tabs.Tab>
            ))
          }
        </Tabs.List>
      </Tabs>

      <Box>
        {activeTab === 'desc' && <div>Description Page</div>}
        <Outlet />
      </Box>
    </Box>
  )
}
