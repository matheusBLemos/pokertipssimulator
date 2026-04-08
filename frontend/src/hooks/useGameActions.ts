import { useCallback } from 'react';
import { gameApi } from '../services/api';
import { useRoomStore } from '../store/roomStore';
import type { ActionType } from '../types';
import toast from 'react-hot-toast';

export function useGameActions() {
  const room = useRoomStore((s) => s.room);
  const setRoom = useRoomStore((s) => s.setRoom);

  const performAction = useCallback(
    async (type: ActionType, amount?: number) => {
      if (!room) return;
      try {
        const updated = await gameApi.performAction(room.id, type, amount);
        setRoom(updated);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Action failed');
      }
    },
    [room, setRoom]
  );

  const startRound = useCallback(async () => {
    if (!room) return;
    try {
      const updated = await gameApi.startRound(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to start round');
    }
  }, [room, setRoom]);

  const advanceStreet = useCallback(async () => {
    if (!room) return;
    try {
      const updated = await gameApi.advanceStreet(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to advance street'
      );
    }
  }, [room, setRoom]);

  const settleRound = useCallback(
    async (winners: { pot_index: number; player_ids: string[] }[]) => {
      if (!room) return;
      try {
        const updated = await gameApi.settleRound(room.id, winners);
        setRoom(updated);
      } catch (err) {
        toast.error(
          err instanceof Error ? err.message : 'Failed to settle round'
        );
      }
    },
    [room, setRoom]
  );

  const pauseGame = useCallback(async () => {
    if (!room) return;
    try {
      const updated = await gameApi.pauseGame(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to toggle pause');
    }
  }, [room, setRoom]);

  const autoSettleRound = useCallback(async () => {
    if (!room) return;
    try {
      const updated = await gameApi.autoSettleRound(room.id);
      setRoom(updated);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to auto-settle round'
      );
    }
  }, [room, setRoom]);

  return { performAction, startRound, advanceStreet, settleRound, pauseGame, autoSettleRound };
}
