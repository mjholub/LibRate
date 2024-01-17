import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';

const searchQueryStore: Writable<string> = writable('');

// client side code needs to get the domain name using window.location
const createWebSocket = (url: string) => {
  const socket = new WebSocket(`wss://${url}/api/search/ws`);

  socket.addEventListener('open', () => {
    socket.send("");
  });

  socket.addEventListener('message', (event) => {
    searchQueryStore.set(event.data);
  });

  return socket;
}

const performSearch = (query: string, socket: WebSocket) => {
  if (socket.readyState === WebSocket.OPEN) {
    socket.send(query);
  }
}

export default {
  subscribe: searchQueryStore.subscribe,
  performSearch,
  createWebSocket
}
