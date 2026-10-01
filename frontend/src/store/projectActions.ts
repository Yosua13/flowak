import type { AppStore } from './useStore';
import { projectsService } from '../services/projects';

type SetStore = (partial: Partial<AppStore> | ((state: AppStore) => Partial<AppStore>)) => void;
type GetStore = () => AppStore;
type ProjectActions = Pick<AppStore,
  'loadProjects' | 'loadArchivedProjects' | 'createProject' | 'deleteProject' | 'restoreProject' |
  'addModule' | 'renameModule' | 'deleteModule'>;

export function createProjectActions(set: SetStore, get: GetStore): ProjectActions {
  return {
    loadProjects: async () => {
      if (!get().token) return;
      try {
        set({ projects: await projectsService.list() });
        await get().loadDashboardStats();
      } catch (error) { console.error('Failed to load projects:', error); }
    },
    loadArchivedProjects: async () => {
      if (!get().token) return;
      try { set({ archivedProjects: await projectsService.list('archived') }); }
      catch (error) { console.error('Failed to load archived projects:', error); }
    },
    createProject: async (name, description) => {
      if (!get().token) return false;
      try {
        await projectsService.create(name, description);
        get().addNotification('Proyek Dibuat', `Proyek "${name}" berhasil ditambahkan.`, 'success');
        await get().loadProjects();
        return true;
      } catch { get().addNotification('Gagal Membuat Proyek', 'Koneksi ke server terputus.', 'warning'); return false; }
    },
    deleteProject: async (id) => {
      if (!get().token) return;
      try {
        await projectsService.archive(id);
        get().addNotification('Proyek Diarsipkan', 'Proyek dapat dipulihkan dari daftar arsip.', 'warning');
        await Promise.all([get().loadProjects(), get().loadArchivedProjects()]);
      } catch (error) { console.error('Failed to archive project:', error); }
    },
    restoreProject: async (id) => {
      if (!get().token) return;
      try {
        await projectsService.restore(id);
        get().addNotification('Proyek Dipulihkan', 'Proyek kembali tersedia di workspace.', 'success');
        await Promise.all([get().loadProjects(), get().loadArchivedProjects()]);
      } catch { get().addNotification('Gagal Memulihkan Proyek', 'Koneksi ke server terputus.', 'warning'); }
    },
    addModule: async (name, description) => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return null;
      try {
        const result = await projectsService.createModule(activeProjectId, { name, description });
        get().addNotification('Modul Ditambahkan', `Modul "${name}" berhasil dibuat.`, 'success');
        set({ modules: await projectsService.listModules(activeProjectId), activeId: result.module_id, selectedNodeId: null });
        return result.module_id;
      } catch (error) { console.error('Failed to add module:', error); return null; }
    },
    renameModule: async (id, name, description) => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return;
      try {
        await projectsService.renameModule(id, name, description);
        set((state) => ({ modules: state.modules.map((module) => module.id === id ? { ...module, name, description: description ?? module.description } : module) }));
      } catch (error) { console.error('Failed to rename module:', error); }
    },
    deleteModule: async (id) => {
      const { token, activeProjectId } = get();
      if (!token || !activeProjectId) return;
      try {
        await projectsService.deleteModule(id);
        get().addNotification('Modul Dihapus', 'Modul berhasil dihapus secara permanen.', 'warning');
        set((state) => {
          const modules = state.modules.filter((module) => module.id !== id);
          return { modules, activeId: modules[0]?.id ?? null, selectedNodeId: null };
        });
      } catch (error) { console.error('Failed to delete module:', error); }
    },
  };
}
