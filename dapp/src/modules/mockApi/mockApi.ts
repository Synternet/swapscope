import { getSubjectName } from '@src/utils';
import { Message } from '../nats';
import { LiquidityPoolItemMessage } from '@src/features';

export async function loadMockedMessages(onMessages: (messages: Message[]) => void) {
  const data = (await import('./data.json')).default as LiquidityPoolItemMessage[];
  const mockedData: Message[] = data.map((x, idx) => {
    return {
      id: idx.toString(),
      data: JSON.stringify(x),
      subject: `${getSubjectName()}.${idx % 2 === 0 ? 'remove' : 'add'}.${x.pair[0].symbol}.${x.pair[1].symbol}`,
      timestamp: 'mockedTimestamp',
    };
  });

  onMessages(mockedData);
}

export function getMockedDateNow() {
  return new Date('2023-10-11T00:00:00.000Z').getTime();
}
