import { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { gameApi } from '../../services/api';
import { useRoomStore } from '../../store/roomStore';
import { useAppStore } from '../../store/appStore';
import { parseToken } from '../../utils/token';
import { isWailsEnvironment, startServer } from '../../services/wailsClient';
import toast from 'react-hot-toast';

export default function GameHomePage() {
  const navigate = useNavigate();
  const setAuth = useRoomStore((s) => s.setAuth);
  const [searchParams] = useSearchParams();
  const codeFromUrl = searchParams.get('code') || '';
  const [mode, setMode] = useState<'menu' | 'create' | 'join'>(codeFromUrl ? 'join' : 'menu');

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-green-400 mb-2">
            Poker With Friends
          </h1>
          <p className="text-gray-400">Full online poker game</p>
        </div>

        {mode === 'menu' && (
          <div className="space-y-4">
            <button
              onClick={() => setMode('create')}
              className="w-full py-4 bg-green-600 hover:bg-green-500 text-white font-semibold rounded-xl text-lg transition-colors"
            >
              Create Room
            </button>
            <button
              onClick={() => setMode('join')}
              className="w-full py-4 bg-gray-700 hover:bg-gray-600 text-white font-semibold rounded-xl text-lg transition-colors"
            >
              Join Room
            </button>
            <button
              onClick={() => navigate('/')}
              className="w-full py-3 text-gray-400 hover:text-white text-sm transition-colors"
            >
              Back to Main Menu
            </button>
          </div>
        )}

        {mode === 'create' && (
          <CreateGameForm
            onBack={() => setMode('menu')}
            onCreated={(token, _roomId, playerId) => {
              setAuth(token, playerId, true);
              navigate('/game/lobby');
            }}
          />
        )}

        {mode === 'join' && (
          <JoinGameForm
            initialCode={codeFromUrl}
            onBack={() => setMode('menu')}
            onJoined={(token, _roomId, playerId) => {
              setAuth(token, playerId, false);
              navigate('/game/lobby');
            }}
          />
        )}
      </div>
    </div>
  );
}

function CreateGameForm({
  onBack,
  onCreated,
}: {
  onBack: () => void;
  onCreated: (token: string, roomId: string, playerId: string) => void;
}) {
  const [name, setName] = useState('');
  const [gameMode, setGameMode] = useState('cash');
  const [stack, setStack] = useState(1000);
  const [port, setPort] = useState(8080);
  const [loading, setLoading] = useState(false);
  const setServerAddress = useAppStore((s) => s.setServerAddress);
  const setConnectionInfo = useAppStore((s) => s.setConnectionInfo);
  const setServerStatus = useAppStore((s) => s.setServerStatus);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Enter your name');
      return;
    }
    setLoading(true);
    try {
      if (isWailsEnvironment()) {
        setServerStatus('starting');
        const connInfo = await startServer(port);
        setConnectionInfo(connInfo);
        setServerAddress(`localhost:${port}`);
        setServerStatus('running');
      }

      const res = await gameApi.createRoom(name.trim(), gameMode, stack, 'game');
      const claims = parseToken(res.token);
      onCreated(res.token, res.room_id, claims.player_id);
    } catch (err) {
      setServerStatus('error');
      toast.error(err instanceof Error ? err.message : 'Failed to create room');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <button
        type="button"
        onClick={onBack}
        className="text-gray-400 hover:text-white text-sm"
      >
        &larr; Back
      </button>

      <div>
        <label className="block text-sm text-gray-400 mb-1">Your Name</label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-green-500"
          placeholder="Enter your name"
          maxLength={20}
          autoFocus
        />
      </div>

      <div>
        <label className="block text-sm text-gray-400 mb-1">Game Mode</label>
        <div className="grid grid-cols-2 gap-2">
          {(['cash', 'tournament'] as const).map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setGameMode(m)}
              className={`py-3 rounded-lg font-medium capitalize transition-colors ${
                gameMode === m
                  ? 'bg-green-600 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {m}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label className="block text-sm text-gray-400 mb-1">
          Starting Stack
        </label>
        <input
          type="number"
          value={stack}
          onChange={(e) => setStack(Number(e.target.value))}
          className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-green-500"
          min={100}
          step={100}
        />
      </div>

      {isWailsEnvironment() && (
        <div>
          <label className="block text-sm text-gray-400 mb-1">
            Server Port
          </label>
          <input
            type="number"
            value={port}
            onChange={(e) => setPort(Number(e.target.value))}
            className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-green-500"
            min={1024}
            max={65535}
          />
        </div>
      )}

      <button
        type="submit"
        disabled={loading}
        className="w-full py-4 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 text-white font-semibold rounded-xl text-lg transition-colors"
      >
        {loading ? 'Creating...' : 'Create Game Room'}
      </button>
    </form>
  );
}

function JoinGameForm({
  initialCode = '',
  onBack,
  onJoined,
}: {
  initialCode?: string;
  onBack: () => void;
  onJoined: (token: string, roomId: string, playerId: string) => void;
}) {
  const [code, setCode] = useState(initialCode);
  const [name, setName] = useState('');
  const [serverAddr, setServerAddr] = useState('');
  const [loading, setLoading] = useState(false);
  const setServerAddress = useAppStore((s) => s.setServerAddress);
  const isWails = isWailsEnvironment();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Enter your name');
      return;
    }
    if (code.length !== 6) {
      toast.error('Room code must be 6 characters');
      return;
    }
    if (isWails && !serverAddr.trim()) {
      toast.error('Enter the server address (IP:Port)');
      return;
    }
    setLoading(true);
    try {
      if (isWails && serverAddr.trim()) {
        setServerAddress(serverAddr.trim());
      }
      const res = await gameApi.joinRoom(code.toUpperCase(), name.trim());
      onJoined(res.token, res.room_id, res.player_id);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to join room');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <button
        type="button"
        onClick={onBack}
        className="text-gray-400 hover:text-white text-sm"
      >
        &larr; Back
      </button>

      <div>
        <label className="block text-sm text-gray-400 mb-1">Your Name</label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-green-500"
          placeholder="Enter your name"
          maxLength={20}
          autoFocus
        />
      </div>

      {isWails && (
        <div>
          <label className="block text-sm text-gray-400 mb-1">
            Server Address
          </label>
          <input
            type="text"
            value={serverAddr}
            onChange={(e) => setServerAddr(e.target.value)}
            className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-green-500"
            placeholder="192.168.1.5:8080"
          />
          <p className="text-xs text-gray-500 mt-1">
            Ask the host for their IP address and port
          </p>
        </div>
      )}

      <div>
        <label className="block text-sm text-gray-400 mb-1">Room Code</label>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase().slice(0, 6))}
          className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white text-center text-2xl tracking-[0.3em] font-mono focus:outline-none focus:border-green-500 uppercase"
          placeholder="ABC123"
          maxLength={6}
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full py-4 bg-green-600 hover:bg-green-500 disabled:bg-gray-700 text-white font-semibold rounded-xl text-lg transition-colors"
      >
        {loading ? 'Joining...' : 'Join Room'}
      </button>
    </form>
  );
}
