import type { AppStore } from './useStore';
import { membersService } from '../services/members';
import { notificationsService } from '../services/notifications';

type SetStore = (partial: Partial<AppStore> | ((state: AppStore) => Partial<AppStore>)) => void;
type GetStore = () => AppStore;
type CollaborationActions = Pick<AppStore,
  'markAllNotificationsRead' | 'loadNotifications' | 'loadTeamMembers' | 'addTeamMember' |
  'deleteTeamMember' | 'loadProjectMembers' | 'addProjectMember' | 'deleteProjectMember' | 'loadDashboardStats'>;

export function createCollaborationActions(set: SetStore, get: GetStore): CollaborationActions {
  return {
    markAllNotificationsRead: () => {
      set((state) => ({ notifications: state.notifications.map((notification) => ({ ...notification, read: true })) }));
      void notificationsService.markAllRead();
    },
    loadNotifications: async () => {
      try { set({ notifications: (await notificationsService.list()).items }); }
      catch { /* Notification availability must not block the workspace. */ }
    },
    loadTeamMembers: async () => {
      if (!get().token) return;
      try { set({ teamMembers: await membersService.listTeam() }); }
      catch (error) { console.error('Failed to load team members:', error); }
    },
    addTeamMember: async (name, email, role) => {
      if (!get().token) return;
      try {
        const data = await membersService.createTeamMember(name, email, role);
        const password = data.temporary_password ? ` Password sementara: ${data.temporary_password}` : '';
        get().addNotification('Anggota Tim Ditambahkan', `${name} dimasukkan ke daftar kontributor.${password}`, 'success');
        await get().loadTeamMembers();
      } catch { get().addNotification('Gagal Menambahkan Anggota', 'Koneksi ke server terputus.', 'warning'); }
    },
    deleteTeamMember: async (id) => {
      const { token, currentUser } = get();
      if (!token) return;
      if (currentUser?.id === id) {
        get().addNotification('Tindakan Dicegah', 'Anda tidak dapat menghapus akun Anda sendiri.', 'warning');
        return;
      }
      try {
        await membersService.deleteTeamMember(id);
        get().addNotification('Anggota Tim Dihentikan', 'Kontributor telah dihapus.', 'warning');
        await get().loadTeamMembers();
      } catch { get().addNotification('Gagal Mengeluarkan Anggota', 'Koneksi ke server terputus.', 'warning'); }
    },
    loadProjectMembers: async () => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return;
      try { set({ projectMembers: await membersService.listProjectMembers(activeProjectId) }); }
      catch (error) { console.error('Failed to load project members:', error); }
    },
    addProjectMember: async (userId) => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return false;
      try {
        await membersService.addProjectMember(activeProjectId, userId);
        get().addNotification('Anggota Ditambahkan', 'Anggota tim berhasil ditambahkan ke proyek.', 'success');
        await get().loadProjectMembers();
        return true;
      } catch { get().addNotification('Gagal Menambahkan Anggota', 'Koneksi ke server terputus.', 'warning'); return false; }
    },
    deleteProjectMember: async (userId) => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return false;
      try {
        await membersService.deleteProjectMember(activeProjectId, userId);
        get().addNotification('Anggota Dihapus', 'Anggota tim telah dihapus dari proyek.', 'warning');
        await get().loadProjectMembers();
        return true;
      } catch { get().addNotification('Gagal Menghapus Anggota', 'Koneksi ke server terputus.', 'warning'); return false; }
    },
    loadDashboardStats: async () => {
      if (!get().token) return;
      try {
        const data = await membersService.dashboard();
        set({ dashboardStats: { myTasksCount: data.my_tasks_count, completionRate: data.completion_rate } });
      } catch (error) { console.error('Failed to load dashboard stats:', error); }
    },
  };
}
