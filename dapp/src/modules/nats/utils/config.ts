import { getNatsUrl } from '@src/utils/env';
import { NatsConfig } from '../types';

export const natsStaticConfig: NatsConfig = {
  url: getNatsUrl(),
};
