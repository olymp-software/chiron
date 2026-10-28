// ImageResolver maps an image src from Markdown to the URL to load, or null to
// show the image's alt text instead.
export type ImageResolver = (src: string) => string | null

// Must match library.FilesPrefix in Go.
const filesPrefix = '/course-files/'

// courseImages resolves image paths relative to baseDir inside the course
// copy with the given storage key. Remote, absolute and escaping paths give
// null, so nothing outside the course is ever loaded.
export function courseImages(key: string, baseDir = ''): ImageResolver {
  return (src) => {
    if (/^[a-z][a-z0-9+.-]*:/i.test(src) || src.startsWith('/') || src.startsWith('\\')) {
      return null
    }
    const path = src.split(/[?#]/)[0]
    const parts = baseDir.split('/').filter((p) => p !== '')
    for (const seg of path.split('/')) {
      if (seg === '..') {
        if (parts.length === 0) {
          return null
        }
        parts.pop()
      } else if (seg !== '.' && seg !== '') {
        parts.push(seg)
      }
    }
    if (parts.length === 0) {
      return null
    }
    return filesPrefix + encodeURIComponent(key) + '/' + parts.join('/')
  }
}
