export function formatChips(amount: number): string {
  if (amount >= 1_000_000) {
    return `${(amount / 1_000_000).toFixed(1)}M`;
  }
  if (amount >= 1_000) {
    return `${(amount / 1_000).toFixed(amount % 1000 === 0 ? 0 : 1)}K`;
  }
  return amount.toString();
}
