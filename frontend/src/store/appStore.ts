import { create } from 'zustand';

type AppMode = 'tips' | 'game';
type ServerStatus = 'stopped' | 'starting' | 'running' | 'error';

export interface ConnectionInfo {
  local_ip: string;
  public_ip: string;
  port: number;
  upnp_ok: boolean;
  local_url: string;
  public_url: string;
}

interface AppState {
  mode: AppMode | null;
  serverStatus: ServerStatus;
  connectionInfo: ConnectionInfo | null;
  serverAddress: string | null;
  setMode: (mode: AppMode) => void;
  setServerStatus: (status: ServerStatus) => void;
  setConnectionInfo: (info: ConnectionInfo) => void;
  setServerAddress: (address: string | null) => void;
  clearMode: () => void;
}

export const useAppStore = create<AppState>((set) => ({
  mode: null,
  serverStatus: 'stopped',
  connectionInfo: null,
  serverAddress: null,

  setMode: (mode) => {
    localStorage.setItem('poker_mode', mode);
    set({ mode });
  },

  setServerStatus: (serverStatus) => set({ serverStatus }),

  setConnectionInfo: (connectionInfo) => set({ connectionInfo }),

  setServerAddress: (serverAddress) => {
    if (serverAddress) {
      localStorage.setItem('poker_server_address', serverAddress);
    } else {
      localStorage.removeItem('poker_server_address');
    }
    set({ serverAddress });
  },

  clearMode: () => {
    localStorage.removeItem('poker_mode');
    localStorage.removeItem('poker_server_address');
    set({ mode: null, serverStatus: 'stopped', connectionInfo: null, serverAddress: null });
  },
}));
