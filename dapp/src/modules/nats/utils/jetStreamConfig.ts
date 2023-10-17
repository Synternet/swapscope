import { isEqual } from 'lodash';
import { NatsConnection, StreamInfo, nanos } from 'nats.ws';
import { JetStreamConfigItem } from '../types';

const config: JetStreamConfigItem[] = [
  {
    streamName: 'swapscope',
    subjects: ['SwapScopeTest.analyticstest0.>'],
    maxAge: nanos(48 * 60 * 60 * 1000),
  },
];

export async function checkJetStreamConfig(connection: NatsConnection) {
  const jetManager = await connection.jetstreamManager();

  const validStreams: string[] = [];
  const streams = await jetManager.streams.list().next();
  for (const stream of streams) {
    const streamName = stream.config.name;
    const configItem = config.find((x) => x.streamName === streamName);

    if (isConfigOutdated(stream, configItem)) {
      await jetManager.streams.delete(streamName);
    } else {
      validStreams.push(streamName);
    }
  }

  for (let i = 0; i < config.length; i++) {
    const item = config[i];
    if (!validStreams.includes(item.streamName)) {
      await jetManager.streams.add({
        name: item.streamName,
        subjects: item.subjects,
        max_age: item.maxAge,
      });

      await jetManager.consumers.add(item.streamName, {});
    }
  }
}

function isConfigOutdated(stream: StreamInfo, configItem?: JetStreamConfigItem) {
  if (!configItem) {
    return true;
  }

  return !isEqual(stream.config.subjects, configItem.subjects) || stream.config.max_age !== configItem?.maxAge;
}

export function getJetStreamName(subject: string) {
  const stream = config.find((x) => x.subjects.includes(subject));
  return stream?.streamName;
}
