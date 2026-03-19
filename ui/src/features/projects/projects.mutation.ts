import { Projects } from '@matrixhub/api-ts/v1alpha1/project.pb'
import { mutationOptions } from '@tanstack/react-query'

import i18n from '@/i18n'

import { projectKeys } from './projects.query'

import type { NotificationMeta } from '@/types/tanstack-query'
import type { ProjectType } from '@matrixhub/api-ts/v1alpha1/project.pb'

interface UpdateProjectInput {
  name: string
  type: ProjectType
}

export function updateProjectMutationOptions() {
  return mutationOptions({
    mutationFn: (input: UpdateProjectInput) =>
      Projects.UpdateProject(input),
    meta: {
      errorMessage: i18n.t('projects.detail.settingsPage.updateError'),
      invalidates: [projectKeys.all],
    } satisfies NotificationMeta,
  })
}
