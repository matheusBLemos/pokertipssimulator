import type { Room } from '../../types';
import { useGameActions } from '../../hooks/useGameActions';
import { STREETS } from '../../utils/constants';

interface HostControlsProps {
  room: Room;
  onSettleClick: () => void;
}

export default function HostControls({ room, onSettleClick }: HostControlsProps) {
  const { advanceStreet, pauseGame, startRound, autoSettleRound } = useGameActions();
  const round = room.round;

  if (!round && room.status === 'waiting') {
    return (
      <div className="bg-gray-900/90 border-t border-gray-800 px-3 py-2 flex gap-2">
        <button
          onClick={startRound}
          className="flex-1 py-2 bg-green-600 hover:bg-green-500 text-white font-medium rounded-lg text-sm transition-colors"
        >
          Start Round
        </button>
      </div>
    );
  }

  if (!round) return null;

  const streetComplete = !round.current_turn && !round.is_complete;
  const isShowdown = round.street === 'showdown' || round.is_complete;

  return (
    <div className="bg-gray-900/90 border-t border-gray-800 px-3 py-2 flex gap-2">
      {streetComplete && !isShowdown && (
        <button
          onClick={advanceStreet}
          className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-lg text-sm transition-colors"
        >
          Next: {STREETS[getNextStreet(round.street)] ?? 'Next'}
        </button>
      )}

      {isShowdown && (
        <>
          <button
            onClick={autoSettleRound}
            className="flex-1 py-2 bg-green-600 hover:bg-green-500 text-white font-medium rounded-lg text-sm transition-colors"
          >
            Auto-Settle
          </button>
          <button
            onClick={onSettleClick}
            className="py-2 px-3 bg-amber-600 hover:bg-amber-500 text-white font-medium rounded-lg text-sm transition-colors"
          >
            Manual
          </button>
        </>
      )}

      <button
        onClick={pauseGame}
        className="py-2 px-4 bg-gray-700 hover:bg-gray-600 text-white font-medium rounded-lg text-sm transition-colors"
      >
        {room.status === 'paused' ? 'Resume' : 'Pause'}
      </button>
    </div>
  );
}

function getNextStreet(current: string): string {
  const order = ['preflop', 'flop', 'turn', 'river', 'showdown'];
  const idx = order.indexOf(current);
  return idx >= 0 && idx < order.length - 1 ? order[idx + 1] : '';
}
