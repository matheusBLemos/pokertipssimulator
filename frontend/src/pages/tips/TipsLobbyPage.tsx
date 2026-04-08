import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { tipsApi } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import { useWebSocket } from '../../hooks/useWebSocket';
import SeatPicker from '../../components/lobby/SeatPicker';
import PlayerList from '../../components/lobby/PlayerList';
import GameSettings from '../../components/lobby/GameSettings';
import ConnectionInfoPanel from '../../components/shared/ConnectionInfo';
import toast from 'react-hot-toast';

export default function TipsLobbyPage() {
  const navigate = useNavigate();
  const { room, token, playerId, isHost, setRoom } = useRoomStore();
  const [loading, setLoading] = useState(true);

  useWebSocket(token);

  useEffect(() => {
    if (!token) {
      navigate('/tips');
      return;
    }

    const claims = JSON.parse(atob(token.split('.')[1]));
    const roomId = claims.room_id;

    tipsApi
      .getRoom(roomId)
      .then(setRoom)
      .catch(() => {
        toast.error('Failed to load room');
        navigate('/tips');
      })
      .finally(() => setLoading(false));
  }, [token, navigate, setRoom]);

  useEffect(() => {
    if (room?.status === 'playing') {
      navigate('/tips/table');
    }
  }, [room?.status, navigate]);

  if (loading || !room) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  const handlePickSeat = async (seat: number) => {
    if (!playerId) return;
    try {
      const updated = await tipsApi.pickSeat(room.id, playerId, seat);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to pick seat');
    }
  };

  const handleStartSession = async () => {
    try {
      const updated = await tipsApi.pauseTimer(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to start session');
    }
  };

  const seatedCount = room.players.filter((p) => p.seat > 0).length;

  return (
    <div className="min-h-screen p-4 max-w-lg mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">Tips Lobby</h2>
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

      {isHost && (
        <ConnectionInfoPanel roomCode={room.code} port={8080} />
      )}

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
            onClick={handleStartSession}
            disabled={seatedCount < 1}
            className="w-full py-4 bg-amber-600 hover:bg-amber-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-xl text-lg transition-colors"
          >
            Start Tips Session
          </button>
        </>
      )}

      <button
        onClick={() => navigate('/tips')}
        className="w-full py-3 text-gray-400 hover:text-white text-sm transition-colors"
      >
        Leave Lobby
      </button>
    </div>
  );
}
