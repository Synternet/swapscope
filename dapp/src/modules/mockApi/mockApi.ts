import { Message } from '../nats';

export async function loadMockedMessages(onMessages: (messages: Message[]) => void) {
  const data = (await import('./data.json')).default;
  const mockedData: Message[] = data.map((x, idx) => ({
    id: idx.toString(),
    data: JSON.stringify(x),
    subject: 'mockedSubject',
    timestamp: 'mockedTimestamp',
  }));

  onMessages(mockedData);
}

export function getMockedDateNow() {
  return new Date('2023-10-11T00:00:00.000Z').getTime();
}
