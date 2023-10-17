import { NatsConnection } from 'nats.ws';

export type NatsMessagesCallback = (messages: Message[]) => void;
export type NatsErrorCallback = (text: string, error: Error) => void;

export interface Message {
  id: string;
  timestamp: string;
  subject: string;
  data: string;
}

export interface NatsConfig {
  url: string;
  connection?: NatsConnection;
  subscription?: { unsubscribe: () => void };
}

export interface NatsWorkerSubscribe {
  type: 'subscribe';
  jwt: string;
  subject: string;
}

export interface NatsWorkerUnsubscribe {
  type: 'unsubscribe';
}

export interface NatsWorkerMessagesReceived {
  type: 'message';
  messages: Message[];
}

export interface NatsWorkerErrorOccurred {
  type: 'error';
  text: string;
  error: Error;
}

export type NatsWorkerReceiveEvents = NatsWorkerSubscribe | NatsWorkerUnsubscribe;
export type NatsWorkerSendEvents = NatsWorkerMessagesReceived | NatsWorkerErrorOccurred;

export interface JetStreamConfigItem {
  streamName: string;
  subjects: string[];
  maxAge: number;
}
