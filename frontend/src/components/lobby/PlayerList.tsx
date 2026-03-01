import type { Player } from '../../types';
import { api } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import { formatChips } from '../../utils/formatChips';
import toast from 'react-hot-toast';

interface PlayerListProps {
  players: Player[];
  hostId: string;
  currentPlayerId: string | null;
  isHost: boolean;
  roomId: string;
}

export default function PlayerList({
  players,
  hostId,
  currentPlayerId,
  isHost,
  roomId,
}: PlayerListProps) {
  const setRoom = useRoomStore((s) => s.setRoom);

  const handleKick = async (playerId: string) => {
    try {
      const updated = await api.kickPlayer(roomId, playerId);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to kick player');
    }
  };

  return (
    <div className="bg-gray-900 rounded-xl p-4">
      <h3 className="text-sm font-medium text-gray-400 mb-3">
        Players ({players.length})
      </h3>
      <div className="space-y-2">
        {players.map((p) => (
          <div
            key={p.id}
            className={`flex items-center justify-between px-3 py-2 rounded-lg ${
              p.id === currentPlayerId ? 'bg-gray-800' : 'bg-gray-800/50'
            }`}
          >
            <div className="flex items-center gap-2">
              <span
                className={`w-2 h-2 rounded-full ${
                  p.seat > 0 ? 'bg-green-500' : 'bg-gray-500'
                }`}
              />
              <span className="text-white font-medium">{p.name}</span>
              {p.id === hostId && (
                <span className="text-xs bg-amber-600/30 text-amber-400 px-1.5 py-0.5 rounded">
                  Host
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              <span className="text-gray-400 text-sm">
                {p.seat > 0 ? `Seat ${p.seat}` : 'Unseated'}
              </span>
              <span className="text-amber-400 text-sm font-mono">
                {formatChips(p.stack)}
              </span>
              {isHost && p.id !== currentPlayerId && (
                <button
                  onClick={() => handleKick(p.id)}
                  className="text-red-400 hover:text-red-300 text-xs"
                >
                  Kick
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
