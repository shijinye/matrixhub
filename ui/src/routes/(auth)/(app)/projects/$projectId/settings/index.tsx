import { createFileRoute } from '@tanstack/react-router'

import { ProjectSettingsPage } from '@/features/projects/pages/ProjectSettingsPage'
import { ensureProjectAccess } from '@/utils/routerAccess'

export const Route = createFileRoute(
  '/(auth)/(app)/projects/$projectId/settings/',
)({
  beforeLoad: async ({ params }) => {
    await ensureProjectAccess(params.projectId)
  },
  component: ProjectSettingsPage,
})
