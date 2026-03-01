import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { api } from '../services/api';
import { useRoomStore } from '../store/roomStore';
import { useWebSocket } from '../hooks/useWebSocket';
import { useGameActions } from '../hooks/useGameActions';
import SeatPicker from '../components/lobby/SeatPicker';
import PlayerList from '../components/lobby/PlayerList';
import GameSettings from '../components/lobby/GameSettings';
import toast from 'react-hot-toast';

export default function LobbyPage() {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { room, token, playerId, isHost, setRoom } = useRoomStore();
  const [loading, setLoading] = useState(true);
  const { startRound } = useGameActions();

  useWebSocket(token);

  useEffect(() => {
    if (!roomId || !token) {
      navigate('/');
      return;
    }

    api
      .getRoom(roomId)
      .then(setRoom)
      .catch(() => {
        toast.error('Failed to load room');
        navigate('/');
      })
      .finally(() => setLoading(false));
  }, [roomId, token, navigate, setRoom]);

  useEffect(() => {
    if (room?.status === 'playing') {
      navigate(`/room/${roomId}/table`);
    }
  }, [room?.status, roomId, navigate]);

  if (loading || !room) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  const handlePickSeat = async (seat: number) => {
    if (!roomId || !playerId) return;
    try {
      const updated = await api.pickSeat(roomId, playerId, seat);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to pick seat');
    }
  };

  const handleStartGame = async () => {
    await startRound();
  };

  const seatedCount = room.players.filter((p) => p.seat > 0).length;

  return (
    <div className="min-h-screen p-4 max-w-lg mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">Lobby</h2>
          <p className="text-gray-400 text-sm">
            Room Code:{' '}
            <span className="font-mono text-amber-400 text-lg tracking-wider">
              {room.code}
            </span>
          </p>
        </div>
        <button
          onClick={() => {
            navigator.clipboard.writeText(room.code);
            toast.success('Code copied!');
          }}
          className="px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg text-sm transition-colors"
        >
          Copy Code
        </button>
      </div>

      <SeatPicker
        maxSeats={room.config.max_players}
        players={room.players}
        currentPlayerId={playerId}
        onPickSeat={handlePickSeat}
      />

      <PlayerList
        players={room.players}
        hostId={room.host_player_id}
        currentPlayerId={playerId}
        isHost={isHost}
        roomId={room.id}
      />

      {isHost && (
        <>
          <GameSettings room={room} />
          <button
            onClick={handleStartGame}
            disabled={seatedCount < 2}
            className="w-full py-4 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-xl text-lg transition-colors"
          >
            {seatedCount < 2
              ? `Need ${2 - seatedCount} more seated players`
              : 'Start Game'}
          </button>
        </>
      )}
    </div>
  );
}
