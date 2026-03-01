interface EliminatedOverlayProps {
  playerName: string;
}

export default function EliminatedOverlay({ playerName }: EliminatedOverlayProps) {
  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50">
      <div className="text-center space-y-4">
        <div className="text-6xl">&#x1F6AB;</div>
        <h2 className="text-2xl font-bold text-red-400">Eliminated</h2>
        <p className="text-gray-400">
          {playerName}, you have been eliminated from the tournament.
        </p>
        <p className="text-gray-500 text-sm">
          You can continue watching the game.
        </p>
      </div>
    </div>
  );
}
