import type { Room } from '../../types';
import PlayerSeat from './PlayerSeat';
import PotDisplay from './PotDisplay';
import { formatChips } from '../../utils/formatChips';
import { STREETS } from '../../utils/constants';

interface PokerTableProps {
  room: Room;
  currentPlayerId: string | null;
}

// Seat positions around an oval table (percentage-based for responsiveness)
const SEAT_POSITIONS: Record<number, { top: string; left: string }> = {
  1: { top: '78%', left: '25%' },
  2: { top: '78%', left: '75%' },
  3: { top: '55%', left: '95%' },
  4: { top: '25%', left: '95%' },
  5: { top: '5%', left: '75%' },
  6: { top: '5%', left: '50%' },
  7: { top: '5%', left: '25%' },
  8: { top: '25%', left: '5%' },
  9: { top: '55%', left: '5%' },
  10: { top: '78%', left: '50%' },
};

export default function PokerTable({ room, currentPlayerId }: PokerTableProps) {
  const round = room.round;

  return (
    <div className="relative flex-1 min-h-0">
      {/* Table felt */}
      <div className="absolute inset-4 rounded-[50%] bg-gradient-to-b from-emerald-800 to-emerald-950 border-4 border-amber-900/60 shadow-[inset_0_4px_30px_rgba(0,0,0,0.5)]">
        {/* Center info */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          {round && (
            <div className="text-center space-y-1">
              <div className="text-xs text-emerald-300/70 uppercase tracking-wider">
                {STREETS[round.street] ?? round.street}
              </div>
              <PotDisplay pots={round.pots} />
              {round.current_bet > 0 && (
                <div className="text-xs text-gray-300">
                  Bet: {formatChips(round.current_bet)}
                </div>
              )}
            </div>
          )}
          {!round && (
            <div className="text-emerald-400/40 text-sm">
              Waiting for round...
            </div>
          )}
        </div>
      </div>

      {/* Player seats */}
      {room.players
        .filter((p) => p.seat > 0)
        .map((player) => {
          const pos = SEAT_POSITIONS[player.seat];
          if (!pos) return null;

          const playerState = round?.player_states?.find(
            (ps) => ps.player_id === player.id
          );

          return (
            <div
              key={player.id}
              className="absolute -translate-x-1/2 -translate-y-1/2"
              style={{ top: pos.top, left: pos.left }}
            >
              <PlayerSeat
                player={player}
                playerState={playerState ?? null}
                isDealer={round?.dealer_seat === player.seat}
                isCurrentTurn={round?.current_turn === player.id}
                isMe={player.id === currentPlayerId}
                isHost={player.id === room.host_player_id}
              />
            </div>
          );
        })}
    </div>
  );
}
