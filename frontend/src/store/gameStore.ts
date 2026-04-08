import { create } from 'zustand';

interface GameState {
  selectedBetAmount: number;
  activeModal: string | null;
  setSelectedBetAmount: (amount: number) => void;
  setActiveModal: (modal: string | null) => void;
  reset: () => void;
}

export const useGameStore = create<GameState>((set) => ({
  selectedBetAmount: 0,
  activeModal: null,

  setSelectedBetAmount: (amount) => set({ selectedBetAmount: amount }),
  setActiveModal: (modal) => set({ activeModal: modal }),
  reset: () => set({ selectedBetAmount: 0, activeModal: null }),
}));
