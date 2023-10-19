import { getNatsUrl } from '@src/utils';
import { NatsConfig } from '../types';

export const natsStaticConfig: NatsConfig = {
  url: getNatsUrl(),
};
