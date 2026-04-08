import { useState } from 'react';
import type { Player } from '../../types';
import { formatChips } from '../../utils/formatChips';
import toast from 'react-hot-toast';

interface ChipTransferProps {
  players: Player[];
  onTransfer: (fromId: string, toId: string, amount: number) => Promise<void>;
}

export default function ChipTransfer({
  players,
  onTransfer,
}: ChipTransferProps) {
  const [fromId, setFromId] = useState('');
  const [toId, setToId] = useState('');
  const [amount, setAmount] = useState(0);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fromId || !toId) {
      toast.error('Select both players');
      return;
    }
    if (fromId === toId) {
      toast.error('Cannot transfer to yourself');
      return;
    }
    if (amount <= 0) {
      toast.error('Enter a valid amount');
      return;
    }

    setLoading(true);
    try {
      await onTransfer(fromId, toId, amount);
      setAmount(0);
      toast.success(`Transferred ${formatChips(amount)} chips`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Transfer failed');
    } finally {
      setLoading(false);
    }
  };

  const seatedPlayers = players.filter((p) => p.seat > 0);

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-gray-900 border border-gray-800 rounded-xl p-4 space-y-3"
    >
      <h3 className="text-sm font-medium text-gray-300">Transfer Chips</h3>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs text-gray-500 mb-1">From</label>
          <select
            value={fromId}
            onChange={(e) => setFromId(e.target.value)}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          >
            <option value="">Select player</option>
            {seatedPlayers.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name} ({formatChips(p.stack)})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs text-gray-500 mb-1">To</label>
          <select
            value={toId}
            onChange={(e) => setToId(e.target.value)}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          >
            <option value="">Select player</option>
            {seatedPlayers
              .filter((p) => p.id !== fromId)
              .map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({formatChips(p.stack)})
                </option>
              ))}
          </select>
        </div>
      </div>

      <div>
        <label className="block text-xs text-gray-500 mb-1">Amount</label>
        <input
          type="number"
          value={amount || ''}
          onChange={(e) => setAmount(Number(e.target.value))}
          className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          placeholder="0"
          min={1}
        />
      </div>

      <button
        type="submit"
        disabled={loading || !fromId || !toId || amount <= 0}
        className="w-full py-2 bg-amber-600 hover:bg-amber-500 disabled:bg-gray-700 disabled:text-gray-500 text-white font-medium rounded-lg text-sm transition-colors"
      >
        {loading ? 'Transferring...' : 'Transfer'}
      </button>
    </form>
  );
}
