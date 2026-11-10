// @vitest-environment jsdom
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import { OpenFolder } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import en from '../i18n/en.json'
import WorkshopBar from './WorkshopBar.vue'

vi.mock('../../wailsjs/go/main/App', () => ({ OpenFolder: vi.fn() }))
vi.mock('../../wailsjs/runtime/runtime', () => ({ ClipboardSetText: vi.fn() }))

beforeEach(() => {
  vi.resetAllMocks()
})

function mountBar() {
  const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } })
  return mount(WorkshopBar, {
    props: { courseKey: 'k8s-1', folder: 'C:\\Users\\me\\Chiron\\k8s-1' },
    global: { plugins: [i18n] },
  })
}

it('shows the folder and copies its path', async () => {
  vi.mocked(ClipboardSetText).mockResolvedValue(true)
  const w = mountBar()

  expect(w.get('code').text()).toBe('C:\\Users\\me\\Chiron\\k8s-1')

  await w.findAll('button')[0].trigger('click')
  await flushPromises()

  expect(ClipboardSetText).toHaveBeenCalledExactlyOnceWith('C:\\Users\\me\\Chiron\\k8s-1')
  expect(w.findAll('button')[0].text()).toBe('Copied')
})

it('opens the workshop folder', async () => {
  vi.mocked(OpenFolder).mockResolvedValue(main.ErrorResult.createFrom({}))
  const w = mountBar()

  await w.findAll('button')[1].trigger('click')
  await flushPromises()

  expect(OpenFolder).toHaveBeenCalledExactlyOnceWith('k8s-1')
  expect(w.find('[role=alert]').exists()).toBe(false)
})

it('shows why the folder did not open', async () => {
  vi.mocked(OpenFolder).mockResolvedValue(
    main.ErrorResult.createFrom({ error: { code: 'workshopFolderFailed' } }),
  )
  const w = mountBar()

  await w.findAll('button')[1].trigger('click')
  await flushPromises()

  expect(w.get('[role=alert]').text()).toContain('Could not open the workshop folder.')
})
