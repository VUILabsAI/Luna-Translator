// Start the community Web app locally, then run:
// node scripts/check-history-browser.mjs http://127.0.0.1:5173
import { readFile } from 'node:fs/promises'
import { chromium, expect } from '@playwright/test'

const baseURL = process.argv[2] || 'http://127.0.0.1:5173'
const browser = await chromium.launch(process.env.PLAYWRIGHT_CHANNEL ? { channel: process.env.PLAYWRIGHT_CHANNEL } : {})
const records = Array.from({ length: 24 }, (_, index) => {
  const number = index + 1
  const id = number.toString(16).padStart(48, '0')
  return {
    id,
    sessionId: `cs-${id}`,
    title: `同传测试记录 ${number}`,
    createdAt: new Date(Date.UTC(2026, 0, 1, 0, 30 - index)).toISOString(),
    provider: 'demo',
    scene: 'quick',
    durationMs: 5000,
    sourceLanguage: 'zh',
    targetLanguage: 'en',
    sourceText: number === 22 ? '' : '第一句。\n\n第二句！',
    translationText: 'First line.\n\nSecond line!',
    units: number === 22 ? [] : [{ id: 'row-1', text: 'First line.', translations: { en: 'First line.', ja: 'こんにちは！' } }],
  }
})

try {
  for (const viewport of [{ width: 1440, height: 900 }, { width: 390, height: 844 }]) {
    const page = await browser.newPage({ baseURL, viewport })
    await page.addInitScript(() => {
      localStorage.setItem('consumer_access_token', 'history-browser-fixture')
    })
    await page.route('**/api/**', async (route) => {
      const { pathname } = new URL(route.request().url())
      const historyMatch = /^\/api\/consumer\/history\/([0-9a-f]+)$/.exec(pathname)
      if (pathname === '/api/config') {
        await route.fulfill({ json: { community: true, demo: true, providers: [], consumerSourceLanguages: ['zh'], consumerTargetLanguages: ['en'] } })
      } else if (pathname === '/api/consumer/history') {
        await route.fulfill({ json: { records } })
      } else if (historyMatch) {
        await route.fulfill({ json: records.find(record => record.id === historyMatch[1]) })
      } else {
        await route.fulfill({ status: 404, json: {} })
      }
    })

    const getScroll = () => page.evaluate(() => Math.max(window.scrollY, document.querySelector('.consumer-page-scroll')?.scrollTop || 0))
    await page.goto('/history')
    const olderRecord = page.getByRole('link', { name: /同传测试记录 21/ })
    await olderRecord.scrollIntoViewIfNeeded()
    const previousScroll = await getScroll()
    expect(previousScroll).toBeGreaterThan(500)
    await olderRecord.click()
    const exportButton = page.getByRole('button', { name: '下载会话文本', exact: true })
    await expect(exportButton).toBeVisible()
    await page.getByLabel('导出译文语言').selectOption('ja')
    const downloaded = page.waitForEvent('download')
    await exportButton.click()
    const text = await readFile(await (await downloaded).path(), 'utf8')
    expect(text).toBe('原文\n第一句。\n\n第二句！\n\n译文（日语）\nこんにちは！')

    await page.getByRole('link', { name: '返回', exact: true }).click()
    await expect(olderRecord).toBeVisible()
    await expect.poll(async () => Math.abs(await getScroll() - previousScroll)).toBeLessThan(5)

    await page.getByRole('link', { name: /同传测试记录 22/ }).click()
    await expect(page.getByLabel('导出译文语言')).toHaveCount(0)
    const translationOnly = page.waitForEvent('download')
    await exportButton.click()
    expect(await readFile(await (await translationOnly).path(), 'utf8')).toBe('译文（英语）\nFirst line.\n\nSecond line!')
    await page.close()
  }
  console.log('History browser checks passed on desktop and mobile-sized Chromium.')
} finally {
  await browser.close()
}
