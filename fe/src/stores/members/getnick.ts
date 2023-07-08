export async function getNick(id: number) {
  const res = await fetch('/api/member/' + id, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  });
  const member = await res.json();
  return res.ok ? member.membername : new Error(`HTTP error! status: ${res.status}`);
};
