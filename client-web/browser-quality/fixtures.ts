import { test as base, expect } from '@playwright/test';

interface BrowserQualityFixtures {
  networkGuard: void;
}

export const test = base.extend<BrowserQualityFixtures>({
  networkGuard: [
    async ({ page, baseURL }, completeFixture) => {
      if (!baseURL) {
        throw new Error('Browser-quality tests require a configured baseURL');
      }

      const applicationOrigin = new URL(baseURL).origin;
      const thirdPartyRequests: string[] = [];
      await page.route('**/*', async (route) => {
        const url = new URL(route.request().url());
        if (url.protocol === 'http:' || url.protocol === 'https:') {
          if (url.origin !== applicationOrigin) {
            thirdPartyRequests.push(url.href);
            await route.abort('blockedbyclient');
            return;
          }
        }
        await route.continue();
      });

      await page.addInitScript(() => {
        window.localStorage.clear();
        window.sessionStorage.clear();
      });
      await completeFixture();
      expect(thirdPartyRequests).toEqual([]);
    },
    { auto: true },
  ],
});

export { expect };
