import { createFileRoute } from '@tanstack/react-router'

import {
  membersQueryOptions,
  membersSearchSchema,
} from '@/features/projects/members/members.query'
import { ProjectMembersPage } from '@/features/projects/members/pages/ProjectMembersPage'

// -- Route definition --

export const Route = createFileRoute(
  '/(auth)/(app)/projects/$projectId/members/',
)({
  validateSearch: membersSearchSchema,
  loaderDeps: ({ search }) => search,
  loader: async ({
    context,
    params,
    deps,
  }) => {
    await context.queryClient.ensureQueryData(
      membersQueryOptions(params.projectId, deps),
    )
  },
  component: ProjectMembersPage,
})
