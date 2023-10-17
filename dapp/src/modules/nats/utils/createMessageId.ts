let key = 0;

export function createMessageId() {
  return `${key++}.${Date.now()}`;
}
