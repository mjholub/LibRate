import axios from "axios";

export const getMaxFileSize = async () => {
  try {
    const response = await axios.get('/api/upload/max-file-size');
    return response.data;
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
