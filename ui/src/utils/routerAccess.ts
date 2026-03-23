import { currentUserQueryOptions, projectRolesQueryOptions } from '@/features/auth/auth.query'
import { projectQueryOptions } from '@/features/projects/projects.query'
import { queryClient } from '@/queryClient'

import type { ProjectRoleType } from '@matrixhub/api-ts/v1alpha1/role.pb'

// ─── Error types ─────────────────────────────────────────────────────────────

export class ForbiddenRouteError extends Error {
  constructor(message = 'You do not have permission to access this page.') {
    super(message)
    this.name = 'ForbiddenRouteError'
    Object.setPrototypeOf(this, ForbiddenRouteError.prototype)
  }
}

export function isForbiddenRouteError(error: unknown): error is ForbiddenRouteError {
  return error instanceof ForbiddenRouteError
    || (error instanceof Error && error.name === 'ForbiddenRouteError')
}

export const forbiddenError = () => new ForbiddenRouteError()

export class NotFoundRouteError extends Error {
  constructor(message = 'The resource you are looking for does not exist.') {
    super(message)
    this.name = 'NotFoundRouteError'
    Object.setPrototypeOf(this, NotFoundRouteError.prototype)
  }
}

export function isNotFoundRouteError(error: unknown): error is NotFoundRouteError {
  return error instanceof NotFoundRouteError
    || (error instanceof Error && error.name === 'NotFoundRouteError')
}

export const notFoundError = () => new NotFoundRouteError()

// ─── Access guard ─────────────────────────────────────────────────────────────

interface EnsureProjectAccessOptions {
  allowPublicRead?: boolean
  allowedRoles?: readonly ProjectRoleType[]
}

export async function ensureProjectAccess(
  projectId: string,
  options?: EnsureProjectAccessOptions,
) {
  const {
    allowPublicRead = false,
    allowedRoles,
  } = options ?? {}

  const [currentUser, projectRoles] = await Promise.all([
    queryClient.ensureQueryData(currentUserQueryOptions()),
    queryClient.ensureQueryData(projectRolesQueryOptions()),
  ])

  if (currentUser.isAdmin) {
    return
  }

  const role = projectRoles.projectRoles?.[projectId]

  if (!role) {
    try {
      await queryClient.ensureQueryData(projectQueryOptions(projectId))
    } catch {
      throw notFoundError()
    }

    // Public projects
    if (allowPublicRead) {
      return
    }

    // 403
    throw forbiddenError()
  }

  if (allowedRoles && !allowedRoles.includes(role)) {
    throw forbiddenError()
  }
}
