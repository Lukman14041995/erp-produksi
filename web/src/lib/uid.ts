// crypto.randomUUID() only exists in "secure contexts" (HTTPS, or
// localhost) -- it throws "crypto.randomUUID is not a function" on plain
// HTTP served from a real IP/domain (verified live: broke every page that
// calls it, e.g. /accounting/journals, on the non-TLS VPS deploy). These
// IDs are only ever used as React list keys -- never sent to the server,
// never relied on for global/cryptographic uniqueness -- so a plain
// fallback works everywhere the real thing doesn't.
export function uid(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}
