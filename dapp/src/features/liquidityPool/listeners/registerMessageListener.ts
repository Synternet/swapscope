import { Message, NatsWorkerSendEvents, NatsWorkerSubscribe, NatsWorkerUnsubscribe, loadMockedMessages } from '@src/modules';
import { createAppJwt } from '@src/modules/nats/utils';
import { getAccessToken, isMockedApi } from '@src/utils';
import { isEqual } from 'lodash';
import { defaultTokenPair, liquidityPoolState, loadData, setLiquidityPoolItems, setTokenPairs } from '../slice';
import { LiquidityPoolItem, TokenPair } from '../types';

export function registerMessageListener(listen: ListenState) {
  listen({
    actionCreator: loadData,
    effect: async (_, api) => {
      const onMessages = (messages: Message[]) => {
        const newItems = messages.map((x) => {
          return { ...JSON.parse(x.data), id: x.id } as LiquidityPoolItem;
        });

        if (newItems.length > 0) {
          const { items, tokenPairs } = liquidityPoolState(api.getState());
          const newList = [...items, ...newItems];
          api.dispatch(setLiquidityPoolItems({ items: newList }));

          if (tokenPairs.length === 1) {
            const tokenPairs = getTopPairs(newItems);
            api.dispatch(setTokenPairs({ tokenPairs }));
          }
        }
      };

      // @TODO handle errors
      const onError = (text: string, error: Error) => {
        console.error(`Nats error: ${text}, ${error}`);
      };

      if (isMockedApi()) {
        loadMockedMessages(onMessages);
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
  const jwt = createAppJwt(getAccessToken());
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

function getTopPairs(list: LiquidityPoolItem[]): TokenPair[] {
  const pairStats: { symbol1: string; symbol2: string; count: number }[] = [];

  list.forEach((item) => {
    const symbol1 = item.pair[0].symbol;
    const symbol2 = item.pair[1].symbol;
    let pair = pairStats.find((x) => x.symbol1 === symbol1 && x.symbol2 === symbol2);
    if (!pair) {
      pair = { symbol1, symbol2, count: 0 };
      pairStats.push(pair);
    }
    pair.count += 1;
  });

  pairStats.sort((a, b) => b.count - a.count);
  const topPairs: TokenPair[] = pairStats
    .map((x) => ({ symbol1: x.symbol1, symbol2: x.symbol2 }))
    .filter((x) => !isEqual(x, defaultTokenPair))
    .slice(0, 4);

  const tokenPairs = [defaultTokenPair, ...topPairs];

  return tokenPairs;
}
