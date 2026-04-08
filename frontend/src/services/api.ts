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
  const res = await fetch(path, {
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

function buildApi(prefix: string) {
  return {
    createRoom: (hostName: string, gameMode: string, startingStack: number, roomMode: string) =>
      request<CreateRoomResponse>(`${prefix}/rooms`, {
        method: 'POST',
        body: JSON.stringify({
          host_name: hostName,
          game_mode: gameMode,
          starting_stack: startingStack,
          room_mode: roomMode,
        }),
      }),

    joinRoom: (code: string, playerName: string) =>
      request<JoinRoomResponse>(`${prefix}/rooms/join`, {
        method: 'POST',
        body: JSON.stringify({ code, player_name: playerName }),
      }),

    getRoom: (roomId: string) =>
      request<Room>(`${prefix}/rooms/${roomId}/`),

    updateConfig: (roomId: string, config: Record<string, unknown>) =>
      request<Room>(`${prefix}/rooms/${roomId}/config`, {
        method: 'PUT',
        body: JSON.stringify(config),
      }),

    pickSeat: (roomId: string, playerId: string, seat: number) =>
      request<Room>(`${prefix}/rooms/${roomId}/players/${playerId}/seat`, {
        method: 'PUT',
        body: JSON.stringify({ seat }),
      }),

    kickPlayer: (roomId: string, playerId: string) =>
      request<Room>(`${prefix}/rooms/${roomId}/players/${playerId}`, {
        method: 'DELETE',
      }),

    rebuy: (roomId: string, playerId: string, amount: number) =>
      request<Room>(`${prefix}/rooms/${roomId}/players/${playerId}/rebuy`, {
        method: 'POST',
        body: JSON.stringify({ amount }),
      }),
  };
}

const sharedGame = buildApi('/api/v1/game');
const sharedTips = buildApi('/api/v1/tips');

export const gameApi = {
  ...sharedGame,

  startRound: (roomId: string) =>
    request<Room>(`/api/v1/game/rooms/${roomId}/rounds/start`, { method: 'POST' }),

  advanceStreet: (roomId: string) =>
    request<Room>(`/api/v1/game/rooms/${roomId}/rounds/advance`, { method: 'POST' }),

  settleRound: (
    roomId: string,
    winners: { pot_index: number; player_ids: string[] }[]
  ) =>
    request<Room>(`/api/v1/game/rooms/${roomId}/rounds/settle`, {
      method: 'POST',
      body: JSON.stringify({ winners }),
    }),

  pauseGame: (roomId: string) =>
    request<Room>(`/api/v1/game/rooms/${roomId}/pause`, { method: 'POST' }),

  performAction: (roomId: string, type: ActionType, amount?: number) =>
    request<Room>(`/api/v1/game/rooms/${roomId}/action`, {
      method: 'POST',
      body: JSON.stringify({ type, amount }),
    }),
};

export const tipsApi = {
  ...sharedTips,

  transferChips: (roomId: string, fromPlayerId: string, toPlayerId: string, amount: number) =>
    request<Room>(`/api/v1/tips/rooms/${roomId}/chips/transfer`, {
      method: 'POST',
      body: JSON.stringify({
        from_player_id: fromPlayerId,
        to_player_id: toPlayerId,
        amount,
      }),
    }),

  advanceBlind: (roomId: string) =>
    request<Room>(`/api/v1/tips/rooms/${roomId}/blinds/advance`, { method: 'POST' }),

  pauseTimer: (roomId: string) =>
    request<Room>(`/api/v1/tips/rooms/${roomId}/pause`, { method: 'POST' }),
};

export const api = gameApi;
