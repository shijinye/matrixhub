import { Box } from '@mantine/core'
import Markdown from 'react-markdown'

import { Route as ModelDetailRoute } from '@/routes/(auth)/(app)/projects_.$projectId/models.$modelId/route'

export function ModelReadmePage() {
  const model = ModelDetailRoute.useLoaderData()

  return (
    <Box pt={20}>
      <Markdown>
        { model?.model.readmeContent }
      </Markdown>
    </Box>
  )
}
