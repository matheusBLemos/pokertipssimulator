import type { WSMessage, Room } from '../types';
import { useRoomStore } from '../store/roomStore';
import toast from 'react-hot-toast';

export function handleWSMessage(msg: WSMessage) {
  const { setRoom } = useRoomStore.getState();

  switch (msg.type) {
    case 'room_state':
    case 'round_started':
    case 'street_advanced':
    case 'settlement':
    case 'round_ended':
      setRoom(msg.payload as Room);
      break;

    case 'player_joined': {
      const room = msg.payload as Room;
      setRoom(room);
      const newPlayer = room.players[room.players.length - 1];
      toast(`${newPlayer?.name ?? 'Someone'} joined the room`);
      break;
    }

    case 'player_left': {
      const payload = msg.payload as { player_id: string };
      const state = useRoomStore.getState();
      if (state.room) {
        const updated = {
          ...state.room,
          players: state.room.players.filter(
            (p) => p.id !== payload.player_id
          ),
        };
        setRoom(updated);
      }
      toast('A player left the room');
      break;
    }

    case 'action_performed': {
      const payload = msg.payload as {
        player_id: string;
        action: string;
        amount: number;
      };
      const state = useRoomStore.getState();
      const player = state.room?.players.find(
        (p) => p.id === payload.player_id
      );
      const name = player?.name ?? 'Player';
      const action = payload.action.toUpperCase();
      const amountStr =
        payload.amount > 0 ? ` ${payload.amount}` : '';
      toast(`${name}: ${action}${amountStr}`, { duration: 2000 });
      break;
    }

    case 'game_paused':
      setRoom(msg.payload as Room);
      toast('Game paused');
      break;

    case 'game_resumed':
      setRoom(msg.payload as Room);
      toast('Game resumed');
      break;

    case 'stack_updated':
    case 'chips_transferred':
      setRoom(msg.payload as Room);
      break;

    case 'blind_level_changed':
      setRoom(msg.payload as Room);
      toast('Blind level advanced');
      break;

    case 'error': {
      const payload = msg.payload as { error: string };
      toast.error(payload.error);
      break;
    }
  }
}
