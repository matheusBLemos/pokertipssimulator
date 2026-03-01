import { useState } from 'react';
import type { Room, PlayerState } from '../../types';
import { useGameActions } from '../../hooks/useGameActions';
import { formatChips } from '../../utils/formatChips';

interface ActionBarProps {
  room: Room;
  isMyTurn: boolean;
  playerState: PlayerState | null;
  playerStack: number;
}

export default function ActionBar({
  room,
  isMyTurn,
  playerState,
  playerStack,
}: ActionBarProps) {
  const { performAction } = useGameActions();
  const [betAmount, setBetAmount] = useState(0);
  const [showSlider, setShowSlider] = useState(false);

  if (!room.round || !isMyTurn || !playerState) {
    return (
      <div className="bg-gray-900 border-t border-gray-800 p-3 text-center text-gray-500 text-sm">
        {isMyTurn ? 'Loading...' : 'Waiting for your turn...'}
      </div>
    );
  }

  const round = room.round;
  const currentBet = round.current_bet;
  const myBet = playerState.bet;
  const callAmount = Math.min(currentBet - myBet, playerStack);
  const minBet = round.big_blind;
  const minRaise = currentBet + round.min_raise;
  const canCheck = currentBet <= myBet;
  const canCall = currentBet > myBet;
  const canBet = currentBet === 0;
  const canRaise = currentBet > 0;

  const handleBetOrRaise = () => {
    if (showSlider) {
      if (canBet) {
        performAction('bet', betAmount);
      } else {
        performAction('raise', betAmount);
      }
      setShowSlider(false);
    } else {
      setBetAmount(canBet ? minBet : minRaise);
      setShowSlider(true);
    }
  };

  const totalPot = round.pots.reduce((s, p) => s + p.amount, 0) +
    round.player_states.reduce((s, ps) => s + ps.bet, 0);

  return (
    <div className="bg-gray-900 border-t border-gray-800 p-3 safe-area-bottom">
      {showSlider && (
        <div className="mb-3 space-y-2">
          <div className="flex items-center gap-2">
            <input
              type="range"
              min={canBet ? minBet : minRaise}
              max={playerStack + myBet}
              value={betAmount}
              onChange={(e) => setBetAmount(Number(e.target.value))}
              className="flex-1 accent-amber-500 h-2"
            />
            <input
              type="number"
              value={betAmount}
              onChange={(e) => setBetAmount(Number(e.target.value))}
              className="w-20 px-2 py-1 bg-gray-800 border border-gray-700 rounded text-white text-sm text-center"
            />
          </div>
          <div className="flex gap-2">
            {[
              { label: '1/2 Pot', value: Math.max(Math.floor(totalPot / 2), canBet ? minBet : minRaise) },
              { label: 'Pot', value: Math.max(totalPot, canBet ? minBet : minRaise) },
              { label: 'All-In', value: playerStack + myBet },
            ].map((preset) => (
              <button
                key={preset.label}
                onClick={() => setBetAmount(Math.min(preset.value, playerStack + myBet))}
                className="flex-1 py-1 bg-gray-800 hover:bg-gray-700 rounded text-xs text-gray-300 transition-colors"
              >
                {preset.label}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="flex gap-2">
        <button
          onClick={() => performAction('fold')}
          className="flex-1 py-3 bg-red-900/80 hover:bg-red-800 text-white font-semibold rounded-xl text-sm transition-colors min-h-[48px]"
        >
          Fold
        </button>

        {canCheck && (
          <button
            onClick={() => performAction('check')}
            className="flex-1 py-3 bg-blue-900/80 hover:bg-blue-800 text-white font-semibold rounded-xl text-sm transition-colors min-h-[48px]"
          >
            Check
          </button>
        )}

        {canCall && (
          <button
            onClick={() => performAction('call')}
            className="flex-1 py-3 bg-blue-900/80 hover:bg-blue-800 text-white font-semibold rounded-xl text-sm transition-colors min-h-[48px]"
          >
            Call {formatChips(callAmount)}
          </button>
        )}

        {(canBet || canRaise) && (
          <button
            onClick={handleBetOrRaise}
            className="flex-1 py-3 bg-amber-600 hover:bg-amber-500 text-white font-semibold rounded-xl text-sm transition-colors min-h-[48px]"
          >
            {showSlider
              ? `${canBet ? 'Bet' : 'Raise'} ${formatChips(betAmount)}`
              : canBet
              ? 'Bet'
              : 'Raise'}
          </button>
        )}

        <button
          onClick={() => performAction('allin')}
          className="py-3 px-4 bg-red-600 hover:bg-red-500 text-white font-semibold rounded-xl text-sm transition-colors min-h-[48px]"
        >
          All-In
        </button>
      </div>
    </div>
  );
}
