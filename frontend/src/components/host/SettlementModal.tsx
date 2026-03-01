import { useState } from 'react';
import type { Room } from '../../types';
import { useGameActions } from '../../hooks/useGameActions';
import { formatChips } from '../../utils/formatChips';

interface SettlementModalProps {
  room: Room;
  onClose: () => void;
}

export default function SettlementModal({ room, onClose }: SettlementModalProps) {
  const { settleRound } = useGameActions();
  const round = room.round;

  const pots = round?.pots ?? [];
  const [selectedWinners, setSelectedWinners] = useState<
    Record<number, string[]>
  >({});

  if (!round) return null;

  const nonFoldedPlayers = room.players.filter((p) => {
    const ps = round.player_states.find((s) => s.player_id === p.id);
    return ps && !ps.folded;
  });

  const toggleWinner = (potIndex: number, playerId: string) => {
    setSelectedWinners((prev) => {
      const current = prev[potIndex] ?? [];
      if (current.includes(playerId)) {
        return {
          ...prev,
          [potIndex]: current.filter((id) => id !== playerId),
        };
      }
      return { ...prev, [potIndex]: [...current, playerId] };
    });
  };

  const handleSettle = async () => {
    const winners = Object.entries(selectedWinners)
      .filter(([, ids]) => ids.length > 0)
      .map(([idx, ids]) => ({
        pot_index: Number(idx),
        player_ids: ids,
      }));

    if (winners.length === 0) return;

    await settleRound(winners);
    onClose();
  };

  const allPotsHaveWinners = pots.every(
    (_, i) => (selectedWinners[i] ?? []).length > 0
  );

  return (
    <div className="fixed inset-0 bg-black/70 flex items-end sm:items-center justify-center z-50">
      <div className="bg-gray-900 w-full max-w-md rounded-t-2xl sm:rounded-2xl p-5 max-h-[80vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-white">Select Winners</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white text-xl"
          >
            &times;
          </button>
        </div>

        {pots.map((pot, potIndex) => (
          <div key={potIndex} className="mb-4">
            <h3 className="text-sm font-medium text-gray-400 mb-2">
              {potIndex === 0 ? 'Main Pot' : `Side Pot ${potIndex}`}:{' '}
              <span className="text-amber-400 font-mono">
                {formatChips(pot.amount)}
              </span>
            </h3>

            <div className="space-y-1">
              {nonFoldedPlayers
                .filter(
                  (p) =>
                    !pot.eligible_ids ||
                    pot.eligible_ids.length === 0 ||
                    pot.eligible_ids.includes(p.id)
                )
                .map((player) => {
                  const isSelected = (
                    selectedWinners[potIndex] ?? []
                  ).includes(player.id);
                  return (
                    <button
                      key={player.id}
                      onClick={() => toggleWinner(potIndex, player.id)}
                      className={`w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors ${
                        isSelected
                          ? 'bg-amber-600/30 border border-amber-500'
                          : 'bg-gray-800 hover:bg-gray-700 border border-transparent'
                      }`}
                    >
                      <span className="text-white font-medium">
                        {player.name}
                      </span>
                      <span className="text-gray-400 text-sm">
                        Seat {player.seat}
                      </span>
                    </button>
                  );
                })}
            </div>

            {(selectedWinners[potIndex] ?? []).length > 1 && (
              <p className="text-xs text-amber-400 mt-1">
                Split pot between {(selectedWinners[potIndex] ?? []).length}{' '}
                players
              </p>
            )}
          </div>
        ))}

        <button
          onClick={handleSettle}
          disabled={!allPotsHaveWinners}
          className="w-full py-3 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-semibold rounded-xl transition-colors"
        >
          Confirm Winners
        </button>
      </div>
    </div>
  );
}
