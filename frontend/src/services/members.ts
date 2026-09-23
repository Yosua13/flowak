import type { ID } from '../domain/types';
import type { TeamMember } from '../config/seedData';
import { apiClient } from './apiClient';
import { queryClient, queryKeys } from './queryClient';

export const membersService = {
  async listTeam(signal?: AbortSignal) { const members = await apiClient.get<TeamMember[]>('/users', signal); queryClient.setQueryData(queryKeys.team, members); return members; },
  createTeamMember(name: string, email: string, role: string) { return apiClient.post<{ temporary_password?: string }>('/users', { name, email, role }); },
  deleteTeamMember(id: ID) { return apiClient.delete(`/users/${id}`); },
  async listProjectMembers(projectId: ID, signal?: AbortSignal) { const members = await apiClient.get<TeamMember[]>(`/projects/${projectId}/members`, signal); queryClient.setQueryData(queryKeys.members(projectId), members); return members; },
  addProjectMember(projectId: ID, userId: ID) { return apiClient.post(`/projects/${projectId}/members`, { user_id: userId }); },
  deleteProjectMember(projectId: ID, userId: ID) { return apiClient.delete(`/projects/${projectId}/members/${userId}`); },
  async dashboard() { const result = await apiClient.get<{ my_tasks_count: number; completion_rate: number }>('/users/dashboard-stats'); queryClient.setQueryData(queryKeys.dashboard, result); return result; },
};
