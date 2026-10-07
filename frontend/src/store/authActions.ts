import type { AppStore } from './useStore';
import { authService } from '../services/authService';
import { sessionLifecycle } from '../services/sessionLifecycle';

type SetStore = (partial: Partial<AppStore> | ((state: AppStore) => Partial<AppStore>)) => void;
type GetStore = () => AppStore;
type AuthActions = Pick<AppStore, 'initializeStore' | 'loginUser' | 'registerUser' | 'logoutUser'>;

async function loadSessionData(get: GetStore) {
  await Promise.all([
    get().loadProjects(),
    get().loadArchivedProjects(),
    get().loadTeamMembers(),
    get().loadNotifications(),
  ]);
}

export function createAuthActions(set: SetStore, get: GetStore): AuthActions {
  return {
    initializeStore: async () => {
      document.documentElement.classList.add('dark');
      if (!sessionLifecycle.hasRefreshHint()) {
        set({ screen: 'login' });
        return;
      }
      try {
        const session = await authService.refresh();
        sessionLifecycle.establish(session.token, session.user);
        set({ token: session.token, currentUser: session.user, organizationRole: session.organization_role, isAuthenticated: true, screen: 'dashboard' });
        await loadSessionData(get);
      } catch {
        sessionLifecycle.clear();
        set({ screen: 'login' });
      }
    },
    loginUser: async (email, password) => {
      try {
        const session = await authService.login(email, password);
        sessionLifecycle.establish(session.token, session.user);
        set({ token: session.token, currentUser: session.user, organizationRole: session.organization_role, isAuthenticated: true, screen: 'dashboard' });
        get().addNotification('Login Sukses', `Selamat datang kembali, ${session.user.name}!`, 'success');
        await loadSessionData(get);
        return true;
      } catch {
        get().addNotification('Gagal Masuk', 'Koneksi ke server terputus.', 'warning');
        return false;
      }
    },
    registerUser: async (name, email, password, role) => {
      try {
        await authService.register(name, email, password, role);
        get().addNotification('Registrasi Sukses', 'Akun berhasil dibuat. Silakan masuk.', 'success');
        set({ screen: 'login' });
        return true;
      } catch {
        get().addNotification('Gagal Mendaftar', 'Koneksi ke server terputus.', 'warning');
        return false;
      }
    },
    logoutUser: () => {
      void authService.logout();
      sessionLifecycle.clear();
      set({
        token: null, currentUser: null, isAuthenticated: false, organizationRole: null, screen: 'login',
        projects: [], archivedProjects: [], activeProjectId: null, modules: [], activeId: null,
        selectedNodeId: null, teamMembers: [], projectMembers: [], dashboardStats: null,
      });
    },
  };
}
