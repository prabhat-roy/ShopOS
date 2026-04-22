import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { ApiKey } from '../types'

export function useApiKeys() {
  return useQuery({ queryKey: ['api-keys'], queryFn: () => api.get<ApiKey[]>('/developer/keys') })
}
export function useCreateKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { name: string; scopes: string[] }) => api.post<{ key: string; apiKey: ApiKey }>('/developer/keys', body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  })
}
export function useRevokeKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/developer/keys/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  })
}
