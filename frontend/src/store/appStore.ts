import { create } from 'zustand';

type AppMode = 'tips' | 'game';

interface AppState {
  mode: AppMode | null;
  serverAddress: string;
  setMode: (mode: AppMode) => void;
  setServerAddress: (addr: string) => void;
  clearMode: () => void;
}

export const useAppStore = create<AppState>((set) => ({
  mode: null,
  serverAddress: '',

  setMode: (mode) => {
    localStorage.setItem('poker_mode', mode);
    set({ mode });
  },

  setServerAddress: (addr) => set({ serverAddress: addr }),

  clearMode: () => {
    localStorage.removeItem('poker_mode');
    set({ mode: null, serverAddress: '' });
  },
}));
