import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, type Router } from 'vue-router'
import { EnterCourse, StartCourse } from '../wailsjs/go/main/App'
import type { main } from '../wailsjs/go/models'
import { createAppRouter, installBackInput } from './router'

vi.mock('../wailsjs/go/main/App', () => ({
  EnterCourse: vi.fn(),
  StartCourse: vi.fn(),
}))

const enterCourse = vi.mocked(EnterCourse)
const startCourse = vi.mocked(StartCourse)

// entered makes EnterCourse return current for every course.
function entered(current: string) {
  enterCourse.mockResolvedValue({ current } as main.EnterResult)
}

let router: Router

beforeEach(() => {
  vi.resetAllMocks()
  router = createAppRouter(createMemoryHistory())
})

describe('entering guard', () => {
  it('enters a course when the key changes', async () => {
    entered('hello')
    await router.push('/library')

    await router.push('/courses/go-1/activities/loops')

    expect(enterCourse).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/loops')
  })

  it('opens the current activity of a course route without a slug', async () => {
    entered('hello')

    await router.push('/courses/go-1')

    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/hello')
  })

  it('does not enter again within the same course', async () => {
    entered('hello')
    await router.push('/courses/go-1')

    await router.push('/courses/go-1/activities/loops')

    expect(enterCourse).toHaveBeenCalledOnce()
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/loops')
  })

  it('enters on switching to another course', async () => {
    entered('hello')
    await router.push('/courses/go-1/activities/loops')

    await router.push('/courses/rust-2')

    expect(enterCourse).toHaveBeenCalledTimes(2)
    expect(enterCourse).toHaveBeenLastCalledWith('rust-2')
    expect(router.currentRoute.value.fullPath).toBe('/courses/rust-2/activities/hello')
  })

  it('opens the current activity again within the same course', async () => {
    entered('hello')
    await router.push('/courses/go-1/activities/loops')

    await router.push('/courses/go-1')

    expect(enterCourse).toHaveBeenCalledOnce()
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/hello')
  })

  it('opens the course overview on error without adding an entry', async () => {
    entered('hello')
    await router.push('/courses/go-1/activities/loops')
    enterCourse.mockResolvedValue({
      current: '',
      error: { code: 'manifestInvalid' },
    } as main.EnterResult)

    await router.push('/courses/rust-2')

    expect(router.currentRoute.value.fullPath).toBe('/library/rust-2')
    entered('hello')
    router.back()
    await vi.waitFor(() =>
      expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/loops'),
    )
  })
})

describe('startup', () => {
  it('enters the course entered last', async () => {
    startCourse.mockResolvedValue('go-1')
    entered('hello')

    await router.push('/')

    expect(enterCourse).toHaveBeenCalledExactlyOnceWith('go-1')
    expect(router.currentRoute.value.fullPath).toBe('/courses/go-1/activities/hello')
  })

  it('opens the library without a course entered last', async () => {
    startCourse.mockResolvedValue('')

    await router.push('/')

    expect(enterCourse).not.toHaveBeenCalled()
    expect(router.currentRoute.value.fullPath).toBe('/library')
  })
})

describe('back input', () => {
  // input returns an event that can be cancelled, with the given fields.
  function input(type: string, fields: object): Event {
    return Object.assign(new Event(type, { cancelable: true }), fields)
  }

  it.each([
    ['mouse back button', input('mouseup', { button: 3 })],
    ['Alt+Left', input('keydown', { altKey: true, key: 'ArrowLeft' })],
  ])('%s goes back', (_, event) => {
    const target = new EventTarget()
    const back = vi.spyOn(router, 'back').mockImplementation(() => {})
    installBackInput(router, target)

    target.dispatchEvent(event)

    expect(back).toHaveBeenCalledOnce()
    expect(event.defaultPrevented).toBe(true)
  })

  it.each([
    ['other mouse button', input('mouseup', { button: 0 })],
    ['Left without Alt', input('keydown', { altKey: false, key: 'ArrowLeft' })],
  ])('%s does not go back', (_, event) => {
    const target = new EventTarget()
    const back = vi.spyOn(router, 'back').mockImplementation(() => {})
    installBackInput(router, target)

    target.dispatchEvent(event)

    expect(back).not.toHaveBeenCalled()
    expect(event.defaultPrevented).toBe(false)
  })
})
