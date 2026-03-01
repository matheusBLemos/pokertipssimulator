export type GameMode = 'cash' | 'tournament';
export type RoomStatus = 'waiting' | 'playing' | 'paused' | 'finished';
export type PlayerStatus = 'waiting' | 'active' | 'sitting_out' | 'eliminated' | 'disconnected';
export type Street = 'preflop' | 'flop' | 'turn' | 'river' | 'showdown';
export type ActionType = 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin';

export interface Player {
  id: string;
  name: string;
  seat: number;
  stack: number;
  status: PlayerStatus;
}

export interface PlayerState {
  player_id: string;
  bet: number;
  total_bet: number;
  has_acted: boolean;
  folded: boolean;
  all_in: boolean;
}

export interface Pot {
  amount: number;
  eligible_ids: string[];
}

export interface Action {
  player_id: string;
  type: ActionType;
  amount: number;
  street: Street;
}

export interface BlindLevel {
  small_blind: number;
  big_blind: number;
  ante: number;
  duration: number;
}

export interface BlindStructure {
  levels: BlindLevel[];
  current_level: number;
}

export interface ChipDenomination {
  value: number;
  color: string;
}

export interface ChipSet {
  denominations: ChipDenomination[];
}

export interface RoomConfig {
  game_mode: GameMode;
  starting_stack: number;
  chip_set: ChipSet;
  blind_structure: BlindStructure;
  max_players: number;
  max_rebuy: number;
}

export interface Round {
  number: number;
  street: Street;
  dealer_seat: number;
  small_blind: number;
  big_blind: number;
  current_turn: string;
  current_bet: number;
  min_raise: number;
  player_states: PlayerState[];
  pots: Pot[];
  actions: Action[];
  is_complete: boolean;
}

export interface Room {
  id: string;
  code: string;
  status: RoomStatus;
  host_player_id: string;
  config: RoomConfig;
  players: Player[];
  round: Round | null;
  round_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateRoomResponse {
  room_id: string;
  code: string;
  token: string;
}

export interface JoinRoomResponse {
  room_id: string;
  player_id: string;
  token: string;
}

export interface WSMessage {
  type: string;
  payload: unknown;
  timestamp: number;
}
