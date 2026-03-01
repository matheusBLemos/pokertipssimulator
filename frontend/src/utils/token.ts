export function parseToken(token: string): {
  room_id: string;
  player_id: string;
  is_host: boolean;
} {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return {
      room_id: decoded.room_id ?? '',
      player_id: decoded.player_id ?? '',
      is_host: decoded.is_host ?? false,
    };
  } catch {
    return { room_id: '', player_id: '', is_host: false };
  }
}
