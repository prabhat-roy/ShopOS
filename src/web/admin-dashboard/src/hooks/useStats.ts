import { useQuery } from '@tanstack/react-query'
import { statsApi } from '../api/client'

export function useStats() {
  return useQuery({ queryKey: ['stats'], queryFn: statsApi.get, refetchInterval: 60_000 })
}
