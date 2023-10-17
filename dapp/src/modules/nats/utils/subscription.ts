import { ConnectionOptions, NatsConnection, connect } from 'nats.ws';
import { NatsConfig, NatsErrorCallback, NatsMessagesCallback } from '../types';
import { createAuthenticator } from './authenticator';
import { decodeJetStreamMessage, decodeMessage } from './codec';
import { natsStaticConfig } from './config';
import { checkJetStreamConfig, getJetStreamName } from './jetStreamConfig';
import { bufferMessages } from './bufferMessages';

async function natsConnect(config: NatsConfig, jwt: string) {
  const options: ConnectionOptions = {
    servers: config.url,
    authenticator: createAuthenticator(jwt),
  };

  return await connect(options);
}

interface NatsSubscribeOptions {
  subject: string;
  connection: NatsConnection;
  onMessages: NatsMessagesCallback;
  onError: NatsErrorCallback;
}

async function natsSubscribe(options: NatsSubscribeOptions) {
  const { subject } = options;
  const streamName = getJetStreamName(subject);
  if (streamName) {
    return jetStreamSubscribe({ ...options, streamName });
  }

  return liveSubscribe(options);
}

function liveSubscribe({ connection, subject, onError, onMessages }: NatsSubscribeOptions) {
  const buffer = bufferMessages(onMessages, { noThrottleCount: 0 });
  const subscription = connection.subscribe(subject, {
    callback: (err, msg) => {
      if (err) {
        onError('NATS subscription error.', err);
        return;
      }

      const message = decodeMessage(msg);
      buffer(message);
    },
  });

  return subscription;
}

interface JetStreamSubscribeOptions extends NatsSubscribeOptions {
  streamName: string;
}

async function jetStreamSubscribe({ connection, onMessages, streamName }: JetStreamSubscribeOptions) {
  const buffer = bufferMessages(onMessages, { noThrottleCount: 0 });
  const jetManager = await connection.jetstreamManager();
  const jetStream = jetManager.jetstream();
  const consumer = await jetStream.consumers.get(streamName);
  const messages = await consumer.consume();
  for await (const msg of messages) {
    const message = decodeJetStreamMessage(msg);
    buffer(message);
  }

  return {
    unsubscribe: () => {
      messages.close();
    },
  };
}

interface SubscribeOptions {
  onMessages: NatsMessagesCallback;
  onError: (text: string, error: Error) => void;
  jwt: string;
  subject: string;
  config?: NatsConfig;
}

export async function subscribe({ onMessages, onError, jwt, subject, config = natsStaticConfig }: SubscribeOptions) {
  if (!config.connection) {
    try {
      config.connection = await natsConnect(config, jwt);
    } catch (err: any) {
      onError('Unable to connect to NATS.', err);
      return;
    }
  }

  await checkJetStreamConfig(config.connection);
  config.subscription = await natsSubscribe({ connection: config.connection, subject, onMessages, onError });
}

export async function unsubscribe(config = natsStaticConfig) {
  if (config.subscription) {
    config.subscription.unsubscribe();
    config.subscription = undefined;
  }

  if (config.connection) {
    config.connection.close();
    config.connection = undefined;
  }
}
