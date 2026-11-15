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
  AddRepoCourse: vi.fn(),
  CancelAdd: vi.fn(),
  ListCourses: vi.fn(),
  PickFolder: vi.fn(),
  RemoteVersions: vi.fn(),
}))

function course(key: string, enrolled: boolean, completed: boolean) {
  return { origin: 'local', source: `/src/${key}`, key, title: key, enrolled, completed }
}

async function mountLibrary(courses: object[]) {
  vi.mocked(ListCourses).mockResolvedValue(main.CoursesResult.createFrom({ courses }))
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: LibraryView }],
  })
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(LibraryView, { global: { plugins: [i18n, router] } })
  await flushPromises()
  return w
}

// sections returns the entries of each library section by its title.
function sections(w: Awaited<ReturnType<typeof mountLibrary>>) {
  return w
    .findAll('.group')
    .map((g) => [g.get('h2').text(), g.findAll('.course').map((c) => c.text())])
}

it('lists courses and workshops in separate sections', async () => {
  const w = await mountLibrary([
    { ...course('k8s', false, false), type: 'workshop' },
    { ...course('go', false, false), type: 'course' },
    course('added-before-workshops', false, false),
  ])

  expect(sections(w)).toEqual([
    ['Courses', ['go', 'added-before-workshops']],
    ['Workshops', ['k8s']],
  ])
})

it('hides a section without entries', async () => {
  const w = await mountLibrary([{ ...course('k8s', false, false), type: 'workshop' }])

  expect(sections(w)).toEqual([['Workshops', ['k8s']]])
})

it('adds from a folder or a repository and names neither kind when empty', async () => {
  const w = await mountLibrary([])

  expect(sections(w)).toEqual([])
  expect(w.findAll('.heading button').map((b) => b.text())).toEqual([
    'Add from folder',
    'Add from repository',
  ])
  expect(w.text()).toContain('Add a course or a workshop')
})

it('opens the repository dialog', async () => {
  HTMLDialogElement.prototype.showModal = vi.fn()
  const w = await mountLibrary([])

  await w.findAll('.heading button')[1].trigger('click')

  expect(w.get('dialog h2').text()).toBe('Add from repository')
})

it('marks a course not on the latest version', async () => {
  const w = await mountLibrary([
    { ...course('old', true, false), origin: 'repo', outdated: true },
    { ...course('latest', true, false), origin: 'repo', outdated: false },
  ])

  const outdated = (title: string) =>
    w
      .findAll('.course')
      .find((c) => c.text().startsWith(title))
      ?.find('.outdated')
  expect(outdated('old')?.text()).toBe('Newer version available')
  expect(outdated('latest')?.exists()).toBe(false)
})

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
