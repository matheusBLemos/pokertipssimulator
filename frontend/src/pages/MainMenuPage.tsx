import { useNavigate } from 'react-router-dom';
import { useAppStore } from '../store/appStore';

export default function MainMenuPage() {
  const navigate = useNavigate();
  const setMode = useAppStore((s) => s.setMode);

  const selectMode = (mode: 'tips' | 'game') => {
    setMode(mode);
    navigate(`/${mode}`);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h1 className="text-5xl font-bold text-amber-400 mb-3">Poker App</h1>
          <p className="text-gray-400 text-lg">Choose your game mode</p>
        </div>

        <div className="space-y-4">
          <button
            onClick={() => selectMode('tips')}
            className="w-full group relative overflow-hidden rounded-2xl border border-gray-700 bg-gray-900 p-6 text-left transition-all hover:border-amber-500/50 hover:bg-gray-800"
          >
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-amber-600/20 text-amber-400 text-2xl">
                $
              </div>
              <div>
                <h2 className="text-xl font-semibold text-white mb-1">
                  Poker Tips
                </h2>
                <p className="text-sm text-gray-400">
                  Digital chip simulator for live games. Track chips, manage
                  blinds, and share a session with friends.
                </p>
              </div>
            </div>
          </button>

          <button
            onClick={() => selectMode('game')}
            className="w-full group relative overflow-hidden rounded-2xl border border-gray-700 bg-gray-900 p-6 text-left transition-all hover:border-green-500/50 hover:bg-gray-800"
          >
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-green-600/20 text-green-400 text-2xl">
                ♠
              </div>
              <div>
                <h2 className="text-xl font-semibold text-white mb-1">
                  Poker With Friends
                </h2>
                <p className="text-sm text-gray-400">
                  Full online poker game. Create a room, invite friends, and
                  play with rounds, actions, and settlements.
                </p>
              </div>
            </div>
          </button>
        </div>
      </div>
    </div>
  );
}
