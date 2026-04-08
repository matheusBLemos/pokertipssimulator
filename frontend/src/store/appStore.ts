import { create } from 'zustand';

type AppMode = 'tips' | 'game';
type ServerStatus = 'stopped' | 'starting' | 'running' | 'error';

interface ConnectionInfo {
  localIP: string;
  publicIP: string;
  port: number;
  upnpOK: boolean;
}

interface AppState {
  mode: AppMode | null;
  serverStatus: ServerStatus;
  connectionInfo: ConnectionInfo | null;
  setMode: (mode: AppMode) => void;
  setServerStatus: (status: ServerStatus) => void;
  setConnectionInfo: (info: ConnectionInfo) => void;
  clearMode: () => void;
}

export const useAppStore = create<AppState>((set) => ({
  mode: null,
  serverStatus: 'stopped',
  connectionInfo: null,

  setMode: (mode) => {
    localStorage.setItem('poker_mode', mode);
    set({ mode });
  },

  setServerStatus: (serverStatus) => set({ serverStatus }),

  setConnectionInfo: (connectionInfo) => set({ connectionInfo }),

  clearMode: () => {
    localStorage.removeItem('poker_mode');
    set({ mode: null, serverStatus: 'stopped', connectionInfo: null });
  },
}));
