import { z } from 'zod'

export const DEFAULT_USERS_PAGE = 1
export const DEFAULT_USERS_PAGE_SIZE = 10

export const usersSearchSchema = z.object({
  page: z.coerce.number().int().positive().optional().catch(DEFAULT_USERS_PAGE),
  query: z.string().trim().optional().catch(undefined),
})

export type UsersSearch = z.infer<typeof usersSearchSchema>
