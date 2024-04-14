import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';

const searchQueryStore: Writable<string> = writable('');

// client side code needs to get the domain name using window.location
const createWebSocket = (url: string) => {
  const provided = new URL(url).origin;
  const current = window.location.origin;

  if (provided !== current) {
    console.error(`Provided URL ${provided} does not match current URL ${current}. WebSocket connection aborted.`);
    return null;
  }

  const socket = new WebSocket(`wss://${provided}/api/search/ws`);

  socket.addEventListener('open', () => {
    socket.send("");
  });

  socket.addEventListener('message', (event) => {
    searchQueryStore.set(event.data);
  });

  return socket;
}

const performSearch = (query: string, socket: WebSocket) => {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    console.error('WebSocket is not connected');
    return;
  }
  socket.send(JSON.stringify({ query }));
}

export default {
  subscribe: searchQueryStore.subscribe,
  performSearch,
  createWebSocket
}
