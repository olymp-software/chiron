// externalLink returns href if it may be opened in the external browser
// (http, https, mailto), or null.
export function externalLink(href: string): string | null {
  return /^(https?:\/\/|mailto:)/i.test(href) ? href : null
}
