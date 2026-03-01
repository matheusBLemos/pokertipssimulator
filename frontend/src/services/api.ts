import { API_BASE } from '../utils/constants';
import type {
  CreateRoomResponse,
  JoinRoomResponse,
  Room,
  ActionType,
} from '../types';

function getToken(): string {
  return localStorage.getItem('poker_token') ?? '';
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getToken()}`,
      ...options.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export const api = {
  createRoom: (hostName: string, gameMode: string, startingStack: number) =>
    request<CreateRoomResponse>('/rooms', {
      method: 'POST',
      body: JSON.stringify({
        host_name: hostName,
        game_mode: gameMode,
        starting_stack: startingStack,
      }),
    }),

  joinRoom: (code: string, playerName: string) =>
    request<JoinRoomResponse>('/rooms/join', {
      method: 'POST',
      body: JSON.stringify({ code, player_name: playerName }),
    }),

  getRoom: (roomId: string) => request<Room>(`/rooms/${roomId}/`),

  updateConfig: (roomId: string, config: Record<string, unknown>) =>
    request<Room>(`/rooms/${roomId}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  pickSeat: (roomId: string, playerId: string, seat: number) =>
    request<Room>(`/rooms/${roomId}/players/${playerId}/seat`, {
      method: 'PUT',
      body: JSON.stringify({ seat }),
    }),

  startRound: (roomId: string) =>
    request<Room>(`/rooms/${roomId}/rounds/start`, { method: 'POST' }),

  advanceStreet: (roomId: string) =>
    request<Room>(`/rooms/${roomId}/rounds/advance`, { method: 'POST' }),

  settleRound: (
    roomId: string,
    winners: { pot_index: number; player_ids: string[] }[]
  ) =>
    request<Room>(`/rooms/${roomId}/rounds/settle`, {
      method: 'POST',
      body: JSON.stringify({ winners }),
    }),

  pauseGame: (roomId: string) =>
    request<Room>(`/rooms/${roomId}/pause`, { method: 'POST' }),

  performAction: (roomId: string, type: ActionType, amount?: number) =>
    request<Room>(`/rooms/${roomId}/action`, {
      method: 'POST',
      body: JSON.stringify({ type, amount }),
    }),

  rebuy: (roomId: string, playerId: string, amount: number) =>
    request<Room>(`/rooms/${roomId}/players/${playerId}/rebuy`, {
      method: 'POST',
      body: JSON.stringify({ amount }),
    }),

  kickPlayer: (roomId: string, playerId: string) =>
    request<Room>(`/rooms/${roomId}/players/${playerId}`, {
      method: 'DELETE',
    }),
};
