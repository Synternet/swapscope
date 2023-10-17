import { defineConfig } from 'cypress';
import { addMatchImageSnapshotPlugin } from 'cypress-image-snapshot/plugin';

// eslint-disable-next-line import/no-default-export
export default defineConfig({
  viewportWidth: 1500,
  viewportHeight: 768,
  e2e: {
    setupNodeEvents(on, config) {
      addMatchImageSnapshotPlugin(on, config);
      setupHeadlessViewport(on);
    },
    chromeWebSecurity: false,
    baseUrl: 'http://localhost:3000',
  },
});

function setupHeadlessViewport(on: any) {
  on('before:browser:launch', (browser: any, launchOptions: any) => {
    if (browser.name === 'chrome' && browser.isHeadless) {
      launchOptions.args.push('--window-size=1500,768');
    }

    if (browser.name === 'electron' && browser.isHeadless) {
      launchOptions.preferences.width = 1500;
      launchOptions.preferences.height = 768;
      launchOptions.preferences.frame = false;
      launchOptions.preferences.useContentSize = true;
    }

    return launchOptions;
  });
}
