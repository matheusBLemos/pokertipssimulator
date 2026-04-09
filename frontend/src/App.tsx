import { HashRouter, Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import MainMenuPage from './pages/MainMenuPage';
import TipsHomePage from './pages/tips/TipsHomePage';
import TipsLobbyPage from './pages/tips/TipsLobbyPage';
import TipsTablePage from './pages/tips/TipsTablePage';
import GameHomePage from './pages/game/GameHomePage';
import GameLobbyPage from './pages/game/GameLobbyPage';
import GameTablePage from './pages/game/GameTablePage';

export default function App() {
  return (
    <HashRouter>
      <Toaster
        position="top-center"
        toastOptions={{
          style: {
            background: '#1f2937',
            color: '#f3f4f6',
            border: '1px solid #374151',
          },
        }}
      />
      <Routes>
        <Route path="/" element={<MainMenuPage />} />

        <Route path="/tips" element={<TipsHomePage />} />
        <Route path="/tips/lobby" element={<TipsLobbyPage />} />
        <Route path="/tips/table" element={<TipsTablePage />} />

        <Route path="/game" element={<GameHomePage />} />
        <Route path="/game/lobby" element={<GameLobbyPage />} />
        <Route path="/game/table" element={<GameTablePage />} />

        <Route
          path="*"
          element={
            <div className="min-h-screen flex items-center justify-center text-gray-400">
              Page not found
            </div>
          }
        />
      </Routes>
    </HashRouter>
  );
}
