import type { Card } from '../../types';

interface PlayingCardProps {
  card?: Card;
  faceDown?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const SUIT_SYMBOLS: Record<string, string> = {
  s: '♠',
  h: '♥',
  d: '♦',
  c: '♣',
};

const SUIT_COLORS: Record<string, string> = {
  s: 'text-gray-900',
  h: 'text-red-600',
  d: 'text-blue-600',
  c: 'text-green-700',
};

const RANK_DISPLAY: Record<string, string> = {
  T: '10',
  J: 'J',
  Q: 'Q',
  K: 'K',
  A: 'A',
};

const SIZE_CLASSES = {
  sm: 'w-7 h-10 text-[10px]',
  md: 'w-9 h-13 text-xs',
  lg: 'w-12 h-16 text-sm',
};

export default function PlayingCard({ card, faceDown, size = 'md' }: PlayingCardProps) {
  const sizeClass = SIZE_CLASSES[size];

  if (faceDown || !card) {
    return (
      <div
        className={`${sizeClass} rounded-md bg-gradient-to-br from-blue-700 to-blue-900 border border-blue-500/50 shadow-md flex items-center justify-center`}
      >
        <div className="w-3/4 h-3/4 rounded-sm border border-blue-400/30 bg-blue-800/50" />
      </div>
    );
  }

  const rankDisplay = RANK_DISPLAY[card.rank] ?? card.rank;
  const suitSymbol = SUIT_SYMBOLS[card.suit] ?? card.suit;
  const suitColor = SUIT_COLORS[card.suit] ?? 'text-gray-900';

  return (
    <div
      className={`${sizeClass} rounded-md bg-white border border-gray-300 shadow-md flex flex-col items-center justify-center leading-tight ${suitColor} font-bold`}
    >
      <span>{rankDisplay}</span>
      <span className="leading-none">{suitSymbol}</span>
    </div>
  );
}
