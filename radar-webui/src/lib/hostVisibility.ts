const STORAGE_KEY = 'radar-visible-host-ids'

/** Saved host row IDs the user chose to show; null = never saved (show all on first visit). */
export function readVisibleHostIds(): string[] | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw == null) return null
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return null
    return parsed.map(String).filter((id) => id.length > 0)
  } catch {
    return null
  }
}

export function writeVisibleHostIds(ids: string[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids))
  } catch {
    /* quota / private mode */
  }
}
