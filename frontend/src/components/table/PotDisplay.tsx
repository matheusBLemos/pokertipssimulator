import type { Pot } from '../../types';
import { formatChips } from '../../utils/formatChips';

interface PotDisplayProps {
  pots: Pot[];
}

export default function PotDisplay({ pots }: PotDisplayProps) {
  const total = pots.reduce((sum, p) => sum + p.amount, 0);

  if (total === 0) return null;

  return (
    <div className="text-center">
      <div className="text-lg font-bold text-amber-400 font-mono">
        {formatChips(total)}
      </div>
      {pots.length > 1 && (
        <div className="flex gap-2 justify-center mt-1">
          {pots.map((pot, i) => (
            <span
              key={i}
              className="text-[10px] bg-gray-900/60 rounded px-1.5 py-0.5 text-gray-300"
            >
              {i === 0 ? 'Main' : `Side ${i}`}: {formatChips(pot.amount)}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
