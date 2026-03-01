import { useState } from 'react';
import type { Room } from '../../types';
import { api } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import toast from 'react-hot-toast';

interface RebuyModalProps {
  room: Room;
  playerId: string;
  onClose: () => void;
}

export default function RebuyModal({ room, playerId, onClose }: RebuyModalProps) {
  const setRoom = useRoomStore((s) => s.setRoom);
  const [amount, setAmount] = useState(room.config.starting_stack);
  const [loading, setLoading] = useState(false);

  const handleRebuy = async () => {
    setLoading(true);
    try {
      const updated = await api.rebuy(room.id, playerId, amount);
      setRoom(updated);
      toast.success(`Added ${amount} chips`);
      onClose();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Rebuy failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div className="bg-gray-900 w-full max-w-sm rounded-2xl p-5">
        <h2 className="text-lg font-bold text-white mb-4">Rebuy</h2>

        <div className="mb-4">
          <label className="block text-sm text-gray-400 mb-1">Amount</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(Number(e.target.value))}
            className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-amber-500"
            min={1}
          />
        </div>

        <div className="flex gap-2">
          <button
            onClick={onClose}
            className="flex-1 py-3 bg-gray-800 hover:bg-gray-700 text-white rounded-xl transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleRebuy}
            disabled={loading || amount <= 0}
            className="flex-1 py-3 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 text-white font-semibold rounded-xl transition-colors"
          >
            {loading ? 'Adding...' : 'Add Chips'}
          </button>
        </div>
      </div>
    </div>
  );
}
