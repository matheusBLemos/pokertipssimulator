import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import HomePage from './pages/HomePage';
import LobbyPage from './pages/LobbyPage';
import TablePage from './pages/TablePage';

export default function App() {
  return (
    <BrowserRouter>
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
        <Route path="/" element={<HomePage />} />
        <Route path="/room/:roomId/lobby" element={<LobbyPage />} />
        <Route path="/room/:roomId/table" element={<TablePage />} />
        <Route
          path="*"
          element={
            <div className="min-h-screen flex items-center justify-center text-gray-400">
              Page not found
            </div>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}
