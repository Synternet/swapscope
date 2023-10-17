import { NatsWorkerErrorOccurred, NatsWorkerMessagesReceived, NatsWorkerReceiveEvents } from './types';
import { subscribe, unsubscribe } from './utils';

addEventListener('message', (event: MessageEvent<NatsWorkerReceiveEvents>) => {
  if (event.data.type === 'subscribe') {
    subscribe({
      jwt: event.data.jwt,
      subject: event.data.subject,
      onError: (text, error) => {
        const message: NatsWorkerErrorOccurred = {
          type: 'error',
          error,
          text,
        };

        postMessage(message);
      },
      onMessages: (messages) => {
        const message: NatsWorkerMessagesReceived = {
          type: 'message',
          messages,
        };

        postMessage(message);
      },
    });
    return;
  }

  if (event.data.type === 'unsubscribe') {
    unsubscribe();
    return;
  }
});
