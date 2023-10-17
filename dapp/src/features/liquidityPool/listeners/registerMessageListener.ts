import { Message, NatsWorkerSendEvents, NatsWorkerSubscribe, NatsWorkerUnsubscribe } from '@src/modules';
import { createAppJwt } from '@src/modules/nats/utils';
import { isMockedApi } from '@src/utils/env';
import exampleData from '../data.json';
import { liquidityPoolState, loadData, setLiquidityPoolItems } from '../slice';
import { LiquidityPoolItem } from '../types';

export function registerMessageListener(listen: ListenState) {
  listen({
    actionCreator: loadData,
    effect: (_, api) => {
      const onMessages = (messages: Message[]) => {
        const newItems = messages.map((x) => {
          return { ...JSON.parse(x.data), id: x.id } as LiquidityPoolItem;
        });

        const filteredItems = newItems.filter((x) => x.pair[0].symbol === 'USDC' && x.pair[1].symbol === 'WETH');
        if (filteredItems.length > 0) {
          const { items } = liquidityPoolState(api.getState());
          api.dispatch(setLiquidityPoolItems({ items: [...items, ...filteredItems] }));
        }
      };

      // @TODO handle errors
      const onError = (text: string, error: Error) => {
        console.error(`Nats error: ${text}, ${error}`);
      };

      if (isMockedApi()) {
        const mockedData: Message[] = exampleData.map((x, idx) => ({
          id: idx.toString(),
          data: JSON.stringify(x),
          subject: 'mockedSubject',
          timestamp: 'mockedTimestamp',
        }));
        onMessages(mockedData);
      } else {
        // @TODO handle disconnect
        connectToNats({ onMessages, onError });
      }
    },
  });
}

interface ConnectToNatsOptions {
  onMessages: (messages: Message[]) => void;
  onError: (text: string, error: Error) => void;
}

function connectToNats({ onMessages, onError }: ConnectToNatsOptions) {
  const worker = new Worker(new URL('../../../modules/nats/worker.ts', import.meta.url));
  // @TODO extract to env variable
  const jwt = createAppJwt('SAALVTY2PLMYU4FHAZTNASYM4CSV5LOQX2NLMM757BCAIFIVIQDUNLQASI');
  const subscribeEvent: NatsWorkerSubscribe = {
    type: 'subscribe',
    jwt,
    subject: 'SwapScopeTest.analyticstest0.>',
  };

  worker.postMessage(subscribeEvent);

  worker.onmessage = (event: MessageEvent<NatsWorkerSendEvents>) => {
    const { data } = event;

    if (data.type === 'message') {
      onMessages(data.messages);
    }

    if (data.type === 'error') {
      onError(data.text, data.error);
    }
  };

  return () => {
    const unsubscribeEvent: NatsWorkerUnsubscribe = {
      type: 'unsubscribe',
    };

    worker.postMessage(unsubscribeEvent);
    worker.terminate();
  };
}
