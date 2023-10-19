export function isMockedApi() {
  return process.env.NEXT_PUBLIC_MOCK_API === 'true';
}

export function getNatsUrl() {
  const NEXT_PUBLIC_NATS_URL = process.env.NEXT_PUBLIC_NATS_URL;
  if (!NEXT_PUBLIC_NATS_URL) {
    throw Error('Missing env variable: NEXT_PUBLIC_NATS_URL');
  }

  return NEXT_PUBLIC_NATS_URL;
}

export function getAccessToken() {
  const NEXT_PUBLIC_ACCESS_TOKEN = process.env.NEXT_PUBLIC_ACCESS_TOKEN;
  if (!NEXT_PUBLIC_ACCESS_TOKEN) {
    throw Error('Missing env variable: NEXT_PUBLIC_ACCESS_TOKEN');
  }

  return NEXT_PUBLIC_ACCESS_TOKEN;
}

type Environment = 'local';

export function getEnv() {
  const NEXT_PUBLIC_ENV = process.env.NEXT_PUBLIC_ENV;
  if (!NEXT_PUBLIC_ENV) {
    throw Error('Missing env variable: NEXT_PUBLIC_ENV');
  }

  return NEXT_PUBLIC_ENV as Environment;
}
