import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { tipsApi } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import { useWebSocket } from '../../hooks/useWebSocket';
import BlindTimer from '../../components/table/BlindTimer';
import ChipTransfer from '../../components/shared/ChipTransfer';
import { formatChips } from '../../utils/formatChips';
import toast from 'react-hot-toast';

export default function TipsTablePage() {
  const navigate = useNavigate();
  const { room, token, isHost, setRoom } = useRoomStore();
  const [loading, setLoading] = useState(!room);

  useWebSocket(token);

  useEffect(() => {
    if (!token) {
      navigate('/tips');
      return;
    }

    if (!room) {
      const claims = JSON.parse(atob(token.split('.')[1]));
      tipsApi
        .getRoom(claims.room_id)
        .then(setRoom)
        .catch(() => {
          toast.error('Failed to load room');
          navigate('/tips');
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [token, room, navigate, setRoom]);

  useEffect(() => {
    if (room?.status === 'waiting' || room?.status === 'finished') {
      navigate('/tips/lobby');
    }
  }, [room?.status, navigate]);

  if (loading || !room) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  const handleTransfer = async (fromId: string, toId: string, amount: number) => {
    const updated = await tipsApi.transferChips(room.id, fromId, toId, amount);
    setRoom(updated);
  };

  const handleAdvanceBlind = async () => {
    try {
      const updated = await tipsApi.advanceBlind(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to advance blind');
    }
  };

  const handlePause = async () => {
    try {
      const updated = await tipsApi.pauseTimer(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to toggle pause');
    }
  };

  const handleRebuy = async (targetPlayerId: string, amount: number) => {
    try {
      const updated = await tipsApi.rebuy(room.id, targetPlayerId, amount);
      setRoom(updated);
      toast.success(`Rebuy: +${formatChips(amount)}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to rebuy');
    }
  };

  const isPaused = room.status === 'paused';
  const isTournament = room.config.game_mode === 'tournament';

  return (
    <div className="min-h-screen flex flex-col bg-gray-950">
      {isTournament && <BlindTimer room={room} />}

      <div className="flex-1 p-4 max-w-lg mx-auto w-full space-y-4 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold text-white">Tips Session</h2>
            <p className="text-gray-400 text-sm">
              Code:{' '}
              <span className="font-mono text-amber-400 tracking-wider">
                {room.code}
              </span>
              {isPaused && (
                <span className="ml-2 text-yellow-500 font-medium">PAUSED</span>
              )}
            </p>
          </div>
        </div>

        <div className="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-gray-300 mb-3">Players</h3>
          <div className="space-y-2">
            {room.players
              .filter((p) => p.seat > 0)
              .sort((a, b) => a.seat - b.seat)
              .map((p) => (
                <div
                  key={p.id}
                  className="flex items-center justify-between py-2 px-3 rounded-lg bg-gray-800/50"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-500 w-6">#{p.seat}</span>
                    <span className="text-white">
                      {p.name}
                      {p.id === room.host_player_id && (
                        <span className="ml-1 text-amber-400 text-xs">(host)</span>
                      )}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-amber-400 font-mono font-medium">
                      {formatChips(p.stack)}
                    </span>
                    {isHost && p.id !== room.host_player_id && (
                      <button
                        onClick={() => handleRebuy(p.id, room.config.starting_stack)}
                        className="text-xs px-2 py-1 bg-green-700 hover:bg-green-600 rounded text-white transition-colors"
                      >
                        +Rebuy
                      </button>
                    )}
                  </div>
                </div>
              ))}
          </div>
        </div>

        <ChipTransfer
          players={room.players}
          onTransfer={handleTransfer}
        />

        {isHost && (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={handlePause}
                className={`py-3 rounded-xl font-medium transition-colors ${
                  isPaused
                    ? 'bg-green-600 hover:bg-green-500 text-white'
                    : 'bg-yellow-600 hover:bg-yellow-500 text-white'
                }`}
              >
                {isPaused ? 'Resume' : 'Pause'}
              </button>
              {isTournament && (
                <button
                  onClick={handleAdvanceBlind}
                  className="py-3 bg-gray-700 hover:bg-gray-600 text-white rounded-xl font-medium transition-colors"
                >
                  Advance Blind
                </button>
              )}
            </div>
          </div>
        )}

        <button
          onClick={() => {
            useRoomStore.getState().clearRoom();
            navigate('/tips');
          }}
          className="w-full py-3 text-gray-400 hover:text-white text-sm transition-colors"
        >
          Leave Session
        </button>
      </div>
    </div>
  );
}
