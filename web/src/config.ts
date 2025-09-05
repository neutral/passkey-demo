export const API_BASE: string = (import.meta.env.VITE_API_BASE as string) || 'http://localhost:8080'

export function apiUrl(path: string): string {
  // Ensure absolute URL against API_BASE
  return new URL(path, API_BASE).toString()
}

