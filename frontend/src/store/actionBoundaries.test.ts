import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { AppStore } from './useStore';

const mocks = vi.hoisted(() => ({
  createProject: vi.fn(),
  listProjects: vi.fn(),
  register: vi.fn(),
}));

vi.mock('../services/projects', () => ({
  projectsService: {
    create: mocks.createProject,
    list: mocks.listProjects,
  },
}));
vi.mock('../services/authService', () => ({
  authService: { register: mocks.register, logout: vi.fn(), refresh: vi.fn(), login: vi.fn() },
}));
vi.mock('../services/sessionLifecycle', () => ({
  sessionLifecycle: { hasRefreshHint: vi.fn(), establish: vi.fn(), clear: vi.fn() },
}));

import { createAuthActions } from './authActions';
import { createProjectActions } from './projectActions';

function harness(overrides: Partial<AppStore> = {}) {
  const notifications: string[] = [];
  const state = {
    token: 'token',
    screen: 'dashboard',
    projects: [],
    addNotification: (title: string) => { notifications.push(title); },
    loadProjects: vi.fn(),
    loadArchivedProjects: vi.fn(),
    loadTeamMembers: vi.fn(),
    loadNotifications: vi.fn(),
    loadDashboardStats: vi.fn(),
    ...overrides,
  } as unknown as AppStore;
  const set = (partial: Partial<AppStore> | ((current: AppStore) => Partial<AppStore>)) => Object.assign(state, typeof partial === 'function' ? partial(state) : partial);
  return { state, set, get: () => state, notifications };
}

describe('store action boundaries', () => {
  beforeEach(() => vi.clearAllMocks());

  it('keeps project refresh orchestration behind the project boundary', async () => {
    const context = harness();
    mocks.createProject.mockResolvedValue({ project_id: 'project-1' });
    const actions = createProjectActions(context.set, context.get);
    await expect(actions.createProject('Traceability', 'Workspace')).resolves.toBe(true);
    expect(mocks.createProject).toHaveBeenCalledWith('Traceability', 'Workspace');
    expect(context.state.loadProjects).toHaveBeenCalledOnce();
    expect(context.notifications).toContain('Proyek Dibuat');
  });

  it('keeps registration state transition behind the auth boundary', async () => {
    const context = harness();
    mocks.register.mockResolvedValue({ success: true });
    const actions = createAuthActions(context.set, context.get);
    await expect(actions.registerUser('Ada', 'ada@example.test', 'secret', 'frontend')).resolves.toBe(true);
    expect(context.state.screen).toBe('login');
    expect(context.notifications).toContain('Registrasi Sukses');
  });
});
