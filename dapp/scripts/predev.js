const fs = require('fs');

const configPath = './.env.local';
const localConfig = {
  NEXT_PUBLIC_ENV: 'local',
};

const content = Object.entries(localConfig)
  .map(([key, value]) => `${key}="${value}"`)
  .join('\n');

if (!fs.existsSync(configPath)) {
  fs.writeFileSync(configPath, content, (err) => {
    if (err) {
      process.exit(1);
    }
  });
}
