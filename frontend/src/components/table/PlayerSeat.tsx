import type { Player, PlayerState } from '../../types';
import { formatChips } from '../../utils/formatChips';

interface PlayerSeatProps {
  player: Player;
  playerState: PlayerState | null;
  isDealer: boolean;
  isCurrentTurn: boolean;
  isMe: boolean;
  isHost: boolean;
}

export default function PlayerSeat({
  player,
  playerState,
  isDealer,
  isCurrentTurn,
  isMe,
}: PlayerSeatProps) {
  const isFolded = playerState?.folded ?? false;
  const isAllIn = playerState?.all_in ?? false;
  const bet = playerState?.bet ?? 0;

  return (
    <div className="relative">
      {/* Dealer button */}
      {isDealer && (
        <div className="absolute -top-1 -right-1 w-5 h-5 bg-white text-black rounded-full flex items-center justify-center text-[10px] font-bold z-10 shadow">
          D
        </div>
      )}

      {/* Player card */}
      <div
        className={`w-20 rounded-lg p-1.5 text-center transition-all ${
          isFolded
            ? 'bg-gray-800/60 opacity-50'
            : isCurrentTurn
            ? 'bg-amber-600/90 ring-2 ring-amber-400 shadow-lg shadow-amber-500/20'
            : isMe
            ? 'bg-blue-900/80 ring-1 ring-blue-500/50'
            : 'bg-gray-800/80'
        }`}
      >
        <div className="text-xs font-medium text-white truncate">
          {player.name}
        </div>
        <div className="text-sm font-bold text-amber-400 font-mono">
          {formatChips(player.stack)}
        </div>
        {isAllIn && (
          <div className="text-[10px] font-bold text-red-400 uppercase">
            All-In
          </div>
        )}
        {isFolded && (
          <div className="text-[10px] text-gray-500 uppercase">Fold</div>
        )}
      </div>

      {/* Current bet */}
      {bet > 0 && !isFolded && (
        <div className="absolute -bottom-4 left-1/2 -translate-x-1/2 bg-gray-900/90 rounded px-1.5 py-0.5 text-[10px] text-amber-300 font-mono whitespace-nowrap">
          {formatChips(bet)}
        </div>
      )}
    </div>
  );
}
