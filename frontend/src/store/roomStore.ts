import { create } from 'zustand';
import type { Room, Player } from '../types';

interface RoomState {
  room: Room | null;
  token: string | null;
  playerId: string | null;
  isHost: boolean;
  setRoom: (room: Room) => void;
  setAuth: (token: string, playerId: string, isHost: boolean) => void;
  clearRoom: () => void;
  currentPlayer: () => Player | null;
}

export const useRoomStore = create<RoomState>((set, get) => ({
  room: null,
  token: null,
  playerId: null,
  isHost: false,

  setRoom: (room) => set({ room }),

  setAuth: (token, playerId, isHost) => {
    localStorage.setItem('poker_token', token);
    localStorage.setItem('poker_player_id', playerId);
    localStorage.setItem('poker_is_host', String(isHost));
    set({ token, playerId, isHost });
  },

  clearRoom: () => {
    localStorage.removeItem('poker_token');
    localStorage.removeItem('poker_player_id');
    localStorage.removeItem('poker_is_host');
    set({ room: null, token: null, playerId: null, isHost: false });
  },

  currentPlayer: () => {
    const { room, playerId } = get();
    if (!room || !playerId) return null;
    return room.players.find((p) => p.id === playerId) ?? null;
  },
}));
