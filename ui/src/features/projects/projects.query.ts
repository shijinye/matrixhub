import { Projects } from '@matrixhub/api-ts/v1alpha1/project.pb'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const projectKeys = {
  all: ['projects'] as const,
  lists: () => [...projectKeys.all, 'list'] as const,
  details: () => [...projectKeys.all, 'detail'] as const,
  detail: (projectId: string) => [...projectKeys.details(), projectId] as const,
}

export function projectDetailQueryOptions(projectId: string) {
  return queryOptions({
    queryKey: projectKeys.detail(projectId),
    queryFn: () => Projects.GetProject({ name: projectId }),
  })
}
// -- Query options factory --

export function projectQueryOptions(projectId: string) {
  return queryOptions({
    queryKey: projectKeys.detail(projectId),
    queryFn: () => Projects.GetProject({ name: projectId }),
  })
}

// -- Custom hook --
export function useProject(projectId: string) {
  return useQuery(projectQueryOptions(projectId))
}
