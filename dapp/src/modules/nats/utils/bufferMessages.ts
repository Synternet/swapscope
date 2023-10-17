import { Message, NatsMessagesCallback } from "../types";

interface Options {
  wait?: number;
  noThrottleCount?: number;
}

export function bufferMessages(callback: NatsMessagesCallback, options?: Options) {
  const { wait = 500, noThrottleCount = 5 } = options || {};

  let buffer: Message[] = [];
  let timeoutId: NodeJS.Timeout | undefined = undefined;
  let count = 0;

  const createTimeout = () => {
    return setTimeout(() => {
      callback(buffer);
      buffer = [];
      timeoutId = undefined;
    }, wait);
  };

  return (msg: Message) => {
    if (count < noThrottleCount) {
      callback([msg]);
      count++;
      return;
    }

    buffer.push(msg);
    if (timeoutId === undefined) {
      timeoutId = createTimeout();
    }
  };
}
