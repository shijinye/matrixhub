import { ProjectType } from '@matrixhub/api-ts/v1alpha1/project.pb'
import { z } from 'zod'

export const updateProjectSchema = z.object({
  type: z.enum([ProjectType.PROJECT_TYPE_PUBLIC, ProjectType.PROJECT_TYPE_PRIVATE]),
})

export type UpdateProjectFormValues = z.infer<typeof updateProjectSchema>
