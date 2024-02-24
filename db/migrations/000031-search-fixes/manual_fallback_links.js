// use this in the browser console/node/bun/whatever if the migration fails
fetch('http://0.0.0.0:5984/genres/_find', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    "selector": {},
    "fields": ["_id", "_rev", "kinds", "name", "descriptions"],
    "limit": 2399,
    "skip": 0,
    "execution_stats": true
  })
})
  .then(response => response.json())
  .then(data => {
    const docs = data.docs;
    docs.forEach(doc => {
      doc.url = `/genres/${doc.kinds[0]}/${doc.name}`;
      fetch(`http://0.0.0.0:5984/genres/${doc._id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(doc)
      })
        .catch(error => console.error('Error:', error));
    });
  })
  .catch((error) => {
    console.error('Error:', error);
  });
