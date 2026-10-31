// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { CourseOverview as LoadOverview, SyncCourse } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import en from '../i18n/en.json'
import CourseOverview from './CourseOverview.vue'

vi.mock('../../wailsjs/go/main/App', () => ({
  CourseOverview: vi.fn(),
  EnrollCourse: vi.fn(),
  SyncCourse: vi.fn(),
}))

const loadOverview = vi.mocked(LoadOverview)

function overview(completed: boolean) {
  return main.OverviewResult.createFrom({
    overview: {
      origin: 'local',
      source: '/src/go',
      key: 'go-1',
      title: 'Go',
      enrolled: true,
      completed,
      text: {},
      check: '',
    },
  })
}

beforeEach(() => {
  vi.resetAllMocks()
})

async function mountOverview() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: { render: () => null } }],
  })
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  const w = mount(CourseOverview, {
    props: { courseKey: 'go-1' },
    global: { plugins: [i18n, router] },
  })
  await flushPromises()
  return w
}

function buttons(w: Awaited<ReturnType<typeof mountOverview>>): string[] {
  return w.findAll('.actions button').map((b) => b.text())
}

it('shows a completed course with Review', async () => {
  loadOverview.mockResolvedValue(overview(true))

  const w = await mountOverview()

  expect(w.get('[role=status]').text()).toBe('Course completed')
  expect(buttons(w)).toEqual(['Sync', 'Review'])
})

it('shows an enrolled course with Continue', async () => {
  loadOverview.mockResolvedValue(overview(false))

  const w = await mountOverview()

  expect(w.find('[role=status]').exists()).toBe(false)
  expect(buttons(w)).toEqual(['Sync', 'Continue'])
})

it('updates the completed status after Sync', async () => {
  loadOverview.mockResolvedValue(overview(true))
  const w = await mountOverview()
  vi.mocked(SyncCourse).mockResolvedValue(main.CourseResult.createFrom({}))
  loadOverview.mockResolvedValue(overview(false))

  await w.get('.actions button').trigger('click')
  await flushPromises()

  expect(w.find('[role=status]').exists()).toBe(false)
  expect(buttons(w)).toEqual(['Sync', 'Continue'])
})
