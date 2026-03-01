import type { Player } from '../../types';

interface SeatPickerProps {
  maxSeats: number;
  players: Player[];
  currentPlayerId: string | null;
  onPickSeat: (seat: number) => void;
}

export default function SeatPicker({
  maxSeats,
  players,
  currentPlayerId,
  onPickSeat,
}: SeatPickerProps) {
  const seatMap = new Map<number, Player>();
  players.forEach((p) => {
    if (p.seat > 0) seatMap.set(p.seat, p);
  });

  const currentSeat = players.find((p) => p.id === currentPlayerId)?.seat ?? 0;

  return (
    <div className="bg-gray-900 rounded-xl p-4">
      <h3 className="text-sm font-medium text-gray-400 mb-3">Pick a Seat</h3>
      <div className="grid grid-cols-5 gap-2">
        {Array.from({ length: maxSeats }, (_, i) => i + 1).map((seat) => {
          const occupant = seatMap.get(seat);
          const isMe = seat === currentSeat;
          const isTaken = !!occupant && !isMe;

          return (
            <button
              key={seat}
              onClick={() => !isTaken && onPickSeat(seat)}
              disabled={isTaken}
              className={`aspect-square rounded-lg flex flex-col items-center justify-center text-xs transition-colors ${
                isMe
                  ? 'bg-amber-600 text-white'
                  : isTaken
                  ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                  : 'bg-gray-800 hover:bg-gray-700 text-gray-300 cursor-pointer'
              }`}
            >
              <span className="font-bold text-sm">{seat}</span>
              {occupant && (
                <span className="truncate w-full text-center px-1">
                  {occupant.name}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
