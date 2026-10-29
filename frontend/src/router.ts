import {
  createRouter,
  createWebHashHistory,
  type RouteLocationNormalized,
  type Router,
  type RouterHistory,
} from 'vue-router'
import { EnterCourse, StartCourse } from '../wailsjs/go/main/App'
import ActivityView from './course/ActivityView.vue'
import CourseScreen from './course/CourseScreen.vue'
import LibraryView from './library/LibraryView.vue'
import CourseOverview from './overview/CourseOverview.vue'

// Empty is the component of routes that only redirect from their guard.
const Empty = { render: () => null }

// courseKey returns the storage key of a route under /courses, or ''.
function courseKey(route: RouteLocationNormalized): string {
  return route.path.startsWith('/courses/') ? String(route.params.key) : ''
}

// createAppRouter creates the router. Navigations the learner starts push;
// a guard returning a location takes the place of the pending navigation, so
// app redirects never add an entry of their own.
export function createAppRouter(history: RouterHistory = createWebHashHistory()): Router {
  // current is the current activity returned by EnterCourse, by storage key.
  const current = new Map<string, string>()

  const router = createRouter({
    history,
    routes: [
      {
        path: '/',
        component: Empty,
        beforeEnter: async () => {
          const key = await StartCourse()
          return key ? `/courses/${key}` : '/library'
        },
      },
      { path: '/library', component: LibraryView },
      {
        path: '/library/:key',
        component: CourseOverview,
        props: (route) => ({ courseKey: route.params.key }),
      },
      {
        path: '/courses/:key',
        component: CourseScreen,
        props: (route) => ({ courseKey: route.params.key }),
        children: [
          // Reached here only within the same course; entering it redirects
          // in the global guard.
          {
            path: '',
            component: Empty,
            beforeEnter: (to) => {
              const key = String(to.params.key)
              return `/courses/${key}/activities/${current.get(key) ?? ''}`
            },
          },
          {
            path: 'activities/:slug',
            component: ActivityView,
            props: (route) => ({ courseKey: route.params.key, slug: route.params.slug }),
          },
        ],
      },
      { path: '/:pathMatch(.*)*', redirect: '/library' },
    ],
  })

  // redirected is the course whose current activity this guard has just
  // redirected to; that navigation still comes from the previous route.
  let redirected = ''

  // Entering a course is decided in Go whenever the course key changes; a
  // per-route guard would not rerun when only :key changes.
  router.beforeEach(async (to, from) => {
    const skip = redirected
    redirected = ''
    const key = courseKey(to)
    if (!key || key === courseKey(from) || key === skip) {
      return true
    }
    const res = await EnterCourse(key)
    if (res.error) {
      return `/library/${key}`
    }
    current.set(key, res.current)
    if (!to.params.slug) {
      redirected = key
      return `/courses/${key}/activities/${res.current}`
    }
    return true
  })

  return router
}

// installBackInput makes the mouse back button and Alt+Left go back in the
// router history instead of leaving it to the webview.
export function installBackInput(router: Router, target: EventTarget = window) {
  const back = (e: Event) => {
    e.preventDefault()
    router.back()
  }
  target.addEventListener('mouseup', (e) => {
    if ((e as MouseEvent).button === 3) {
      back(e)
    }
  })
  target.addEventListener('keydown', (e) => {
    const k = e as KeyboardEvent
    if (k.altKey && k.key === 'ArrowLeft') {
      back(e)
    }
  })
}
