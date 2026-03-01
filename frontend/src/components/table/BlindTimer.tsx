import { useEffect, useState } from 'react';
import type { Room } from '../../types';
import { formatChips } from '../../utils/formatChips';

interface BlindTimerProps {
  room: Room;
}

export default function BlindTimer({ room }: BlindTimerProps) {
  const { blind_structure } = room.config;
  const currentLevel = blind_structure.levels[blind_structure.current_level];
  const nextLevel = blind_structure.levels[blind_structure.current_level + 1];
  const [timeLeft, setTimeLeft] = useState(currentLevel?.duration ?? 0);

  useEffect(() => {
    if (!currentLevel?.duration) return;
    setTimeLeft(currentLevel.duration);

    const interval = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) return currentLevel.duration;
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [blind_structure.current_level, currentLevel?.duration]);

  if (!currentLevel) return null;

  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;

  return (
    <div className="bg-gray-900/90 border-b border-gray-800 px-4 py-2 flex items-center justify-between text-sm">
      <div className="flex items-center gap-3">
        <span className="text-gray-400">Level {blind_structure.current_level + 1}</span>
        <span className="text-amber-400 font-mono">
          {formatChips(currentLevel.small_blind)}/{formatChips(currentLevel.big_blind)}
        </span>
        {currentLevel.ante > 0 && (
          <span className="text-gray-500">Ante: {formatChips(currentLevel.ante)}</span>
        )}
      </div>

      <div className="flex items-center gap-3">
        {currentLevel.duration > 0 && (
          <span className="text-white font-mono">
            {minutes}:{seconds.toString().padStart(2, '0')}
          </span>
        )}
        {nextLevel && (
          <span className="text-gray-500 text-xs">
            Next: {formatChips(nextLevel.small_blind)}/{formatChips(nextLevel.big_blind)}
          </span>
        )}
      </div>
    </div>
  );
}
