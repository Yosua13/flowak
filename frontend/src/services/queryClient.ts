import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1, refetchOnWindowFocus: false } },
});

export const queryKeys = {
  projects: ['projects'] as const,
  project: (projectId: string) => ['projects', projectId] as const,
  members: (projectId: string) => ['projects', projectId, 'members'] as const,
  team: ['team'] as const,
  dashboard: ['dashboard'] as const,
};

export async function invalidateProjectCache(projectId?: string) {
  await queryClient.invalidateQueries({ queryKey: projectId ? queryKeys.project(projectId) : queryKeys.projects });
  if (projectId) await queryClient.invalidateQueries({ queryKey: queryKeys.projects });
}
