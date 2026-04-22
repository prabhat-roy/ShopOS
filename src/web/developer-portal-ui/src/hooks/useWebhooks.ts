import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Webhook } from '../types'

export function useWebhooks() {
  return useQuery({ queryKey: ['webhooks'], queryFn: () => api.get<Webhook[]>('/developer/webhooks') })
}
export function useCreateWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { url: string; events: string[] }) => api.post<Webhook>('/developer/webhooks', body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhooks'] }),
  })
}
export function useDeleteWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/developer/webhooks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhooks'] }),
  })
}
