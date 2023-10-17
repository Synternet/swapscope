import { JsMsg, Msg, StringCodec } from 'nats.ws';
import { Message } from '../types';
import * as codec from './codec';
import { createMessageId } from './createMessageId';

const sc = StringCodec();

export function decode(data: Uint8Array) {
  return sc.decode(data);
}

export function decodeMessage(message: Msg): Message {
  const data = codec.decode(message.data);
  return { id: createMessageId(), timestamp: new Date().toISOString(), subject: message.subject, data };
}

export function decodeJetStreamMessage(message: JsMsg): Message {
  return {
    data: decode(message.data),
    id: createMessageId(),
    subject: message.subject,
    timestamp: new Date(message.info.timestampNanos / 1000000).toISOString(),
  };
}
