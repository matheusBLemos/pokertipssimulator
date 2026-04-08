import { useState } from 'react';
import type { Room } from '../../types';
import { getApiForMode } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import toast from 'react-hot-toast';

interface GameSettingsProps {
  room: Room;
}

export default function GameSettings({ room }: GameSettingsProps) {
  const setRoom = useRoomStore((s) => s.setRoom);
  const [sb, setSb] = useState(
    room.config.blind_structure.levels[0]?.small_blind ?? 5
  );
  const [bb, setBb] = useState(
    room.config.blind_structure.levels[0]?.big_blind ?? 10
  );

  const handleSave = async () => {
    try {
      const updated = await getApiForMode(room.mode).updateConfig(room.id, {
        blind_structure: {
          levels: [{ small_blind: sb, big_blind: bb, ante: 0, duration: 0 }],
          current_level: 0,
        },
      });
      setRoom(updated);
      toast.success('Settings updated');
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to update settings'
      );
    }
  };

  return (
    <div className="bg-gray-900 rounded-xl p-4">
      <h3 className="text-sm font-medium text-gray-400 mb-3">Game Settings</h3>

      <div className="grid grid-cols-2 gap-3 mb-3">
        <div>
          <label className="block text-xs text-gray-500 mb-1">
            Small Blind
          </label>
          <input
            type="number"
            value={sb}
            onChange={(e) => setSb(Number(e.target.value))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
            min={1}
          />
        </div>
        <div>
          <label className="block text-xs text-gray-500 mb-1">Big Blind</label>
          <input
            type="number"
            value={bb}
            onChange={(e) => setBb(Number(e.target.value))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
            min={1}
          />
        </div>
      </div>

      <div className="flex items-center justify-between text-sm text-gray-400 mb-3">
        <span>Mode: {room.config.game_mode}</span>
        <span>Stack: {room.config.starting_stack}</span>
      </div>

      <button
        onClick={handleSave}
        className="w-full py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg text-sm transition-colors"
      >
        Save Settings
      </button>
    </div>
  );
}
