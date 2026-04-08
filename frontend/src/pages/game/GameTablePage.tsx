import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { gameApi } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import { useWebSocket } from '../../hooks/useWebSocket';
import PokerTable from '../../components/table/PokerTable';
import ActionBar from '../../components/table/ActionBar';
import BlindTimer from '../../components/table/BlindTimer';
import HostControls from '../../components/host/HostControls';
import SettlementModal from '../../components/host/SettlementModal';
import RebuyModal from '../../components/host/RebuyModal';
import EliminatedOverlay from '../../components/table/EliminatedOverlay';
import toast from 'react-hot-toast';

export default function GameTablePage() {
  const navigate = useNavigate();
  const { room, token, playerId, isHost, setRoom } = useRoomStore();
  const [loading, setLoading] = useState(!room);
  const [showSettle, setShowSettle] = useState(false);
  const [showRebuy, setShowRebuy] = useState(false);

  useWebSocket(token);

  useEffect(() => {
    if (!token) {
      navigate('/game');
      return;
    }

    if (!room) {
      const claims = JSON.parse(atob(token.split('.')[1]));
      gameApi
        .getRoom(claims.room_id)
        .then(setRoom)
        .catch(() => {
          toast.error('Failed to load room');
          navigate('/game');
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [token, room, navigate, setRoom]);

  useEffect(() => {
    if (room?.status === 'waiting' && !room.round) {
      navigate('/game/lobby');
    }
  }, [room?.status, room?.round, navigate]);

  if (loading || !room) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  const isMyTurn = room.round?.current_turn === playerId;
  const myState = room.round?.player_states?.find(
    (ps) => ps.player_id === playerId
  );
  const myPlayer = room.players.find((p) => p.id === playerId);
  const isCashGame = room.config.game_mode === 'cash';
  const isTournament = room.config.game_mode === 'tournament';
  const isEliminated = myPlayer?.status === 'eliminated';

  return (
    <div className="h-screen flex flex-col bg-gray-950">
      {isTournament && <BlindTimer room={room} />}

      <PokerTable room={room} currentPlayerId={playerId} />

      {isHost && (
        <HostControls
          room={room}
          onSettleClick={() => setShowSettle(true)}
        />
      )}

      {myPlayer && room.round && !room.round.is_complete && !isEliminated && (
        <ActionBar
          room={room}
          isMyTurn={isMyTurn}
          playerState={myState ?? null}
          playerStack={myPlayer.stack}
        />
      )}

      {isCashGame && myPlayer && !room.round && (
        <div className="bg-gray-900 border-t border-gray-800 p-3 flex justify-center">
          <button
            onClick={() => setShowRebuy(true)}
            className="px-6 py-2 bg-green-700 hover:bg-green-600 text-white rounded-lg text-sm transition-colors"
          >
            Rebuy
          </button>
        </div>
      )}

      {showSettle && room.round && (
        <SettlementModal
          room={room}
          onClose={() => setShowSettle(false)}
        />
      )}

      {showRebuy && playerId && (
        <RebuyModal
          room={room}
          playerId={playerId}
          onClose={() => setShowRebuy(false)}
        />
      )}

      {isEliminated && myPlayer && (
        <EliminatedOverlay playerName={myPlayer.name} />
      )}
    </div>
  );
}
