// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { ListCourses } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import LibraryView from './LibraryView.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  AddLocalCourse: vi.fn(),
  ListCourses: vi.fn(),
  PickFolder: vi.fn(),
}))

function course(key: string, enrolled: boolean, completed: boolean) {
  return { origin: 'local', source: `/src/${key}`, key, title: key, enrolled, completed }
}

it('marks enrolled and completed courses', async () => {
  vi.mocked(ListCourses).mockResolvedValue(
    main.CoursesResult.createFrom({
      courses: [
        course('added', false, false),
        course('enrolled', true, false),
        course('done', true, true),
      ],
    }),
  )
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: LibraryView }],
  })
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })

  const w = mount(LibraryView, { global: { plugins: [i18n, router] } })
  await flushPromises()

  const status = (title: string) =>
    w
      .findAll('.course')
      .find((c) => c.text().startsWith(title))
      ?.find('.status')
  expect(status('added')?.exists()).toBe(false)
  expect(status('enrolled')?.text()).toBe('Enrolled')
  expect(status('done')?.text()).toBe('Completed')
})
