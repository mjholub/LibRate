
export const getMaxFileSize = async () => {
  try {
    const response = await fetch('/api/upload/max-file-size',
      { method: 'GET' });
    if (!response.ok) {
      return 0;
    }
    const data = await response.text();
    return parseInt(data, 10);
  } catch (error) {
    return 0;
  }
};

type EventHandlerFunc = (event: Event) => void;

// accept is the MIME type of the file
export const openFilePicker = (handler: EventHandlerFunc, accept: string) => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = accept;
  input.addEventListener('change', handler);
  input.click();
};
