import {
  Button, Group, Radio, rem, Stack, Text, TextInput,
} from '@mantine/core'
import { ProjectType } from '@matrixhub/api-ts/v1alpha1/project.pb'
import { useForm } from '@tanstack/react-form'
import { useMutation, useSuspenseQuery } from '@tanstack/react-query'
import { getRouteApi } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { updateProjectMutationOptions } from '../projects.mutation'
import { projectDetailQueryOptions } from '../projects.query'

const projectRouteApi = getRouteApi('/(auth)/(app)/projects/$projectId')

export function ProjectSettingsPage() {
  const { t } = useTranslation()
  const { projectId } = projectRouteApi.useParams()
  const { data: project } = useSuspenseQuery(projectDetailQueryOptions(projectId))
  const mutation = useMutation(updateProjectMutationOptions())

  const isProxy = !!project.registryUrl

  const form = useForm({
    defaultValues: {
      type: project.type ?? ProjectType.PROJECT_TYPE_PRIVATE,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({
        name: projectId,
        type: value.type,
      })
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
    >
      <Stack pt="lg" gap="md" align="flex-start">
        <Stack gap="xs">
          <Text fw={600} fz="sm" lh={rem('20px')} c="var(--mantine-color-gray-7)">
            {t('projects.detail.settingsPage.projectType')}
          </Text>
          <form.Field name="type">
            {field => (
              <Radio.Group
                value={field.state.value}
                onChange={value => field.handleChange(value as ProjectType)}
              >
                <Group gap="xl">
                  <Radio
                    value={ProjectType.PROJECT_TYPE_PUBLIC}
                    label={t('projects.detail.settingsPage.public')}
                  />
                  <Radio
                    value={ProjectType.PROJECT_TYPE_PRIVATE}
                    label={t('projects.detail.settingsPage.private')}
                  />
                </Group>
              </Radio.Group>
            )}
          </form.Field>
        </Stack>

        {isProxy && (
          <Stack gap="xs" w="100%">
            <Text fw={600} fz="sm" lh="20px" c="var(--mantine-color-gray-7)">
              {t('projects.detail.settingsPage.proxyAddress')}
            </Text>
            <TextInput
              value={project.registryUrl ?? ''}
              w={rem('206px')}
              disabled
            />
          </Stack>
        )}

        <Group gap="md">
          <form.Subscribe selector={s => [s.isSubmitting]}>
            {([isSubmitting]) => (
              <>
                <Button
                  type="submit"
                  loading={isSubmitting}
                >
                  {t('projects.detail.settingsPage.save')}
                </Button>
                <Button
                  variant="white"
                  c="gray.9"
                  onClick={() => form.reset()}
                >
                  {t('projects.detail.settingsPage.cancel')}
                </Button>
              </>
            )}
          </form.Subscribe>
        </Group>
      </Stack>
    </form>
  )
}
