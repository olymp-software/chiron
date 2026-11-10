// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter, isNavigationFailure } from 'vue-router'
import { CourseTree, ListCourses } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import CourseScreen from './CourseScreen.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  CourseTree: vi.fn(),
  ListCourses: vi.fn(),
}))

it('shows the course on the screen again when the learner stays', async () => {
  vi.mocked(CourseTree).mockResolvedValue(
    main.TreeResult.createFrom({
      tree: { title: 'Go', type: 'course', current: 'hello', completed: false, sections: [] },
    }),
  )
  const entry = (key: string) => ({ origin: 'local', source: key, key, title: key, enrolled: true })
  vi.mocked(ListCourses).mockResolvedValue(
    main.CoursesResult.createFrom({ courses: [entry('go-1'), entry('k8s-1')] }),
  )
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/courses/:key', component: { render: () => null } }],
  })
  await router.push('/courses/go-1')
  // A declined leave confirmation keeps the learner where they are.
  router.beforeEach((to) => to.params.key !== 'k8s-1')
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(CourseScreen, {
    props: { courseKey: 'go-1' },
    global: { plugins: [i18n, router], stubs: { RouterView: true, CourseTree: true } },
  })
  await flushPromises()
  const push = vi.spyOn(router, 'push')

  await w.get('select').setValue('k8s-1')
  await flushPromises()

  expect(isNavigationFailure(await push.mock.results[0].value)).toBe(true)
  expect(router.currentRoute.value.params.key).toBe('go-1')
  expect((w.get('select').element as HTMLSelectElement).value).toBe('go-1')
})
