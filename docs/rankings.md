# 榜单维护

README 首屏的两张榜单卡片和「项目状态 › 榜单记录」表格都由 [`scripts/readme-badges.mjs`](../scripts/readme-badges.mjs) 根据 [`rankings.json`](rankings.json) 生成。不要直接改 README 里 `<!-- rankings:… -->` 标记之间的内容，下次生成会被覆盖。

```sh
node scripts/readme-badges.mjs           # 按 rankings.json 重新生成卡片、按钮和两份 README
node scripts/readme-badges.mjs refresh   # 先查 GitHub 主题榜和 GitHub Trending，再生成
node scripts/readme-badges.mjs --check   # 只检查是否需要重新生成（CI 会跑）
```

`refresh` 不需要令牌；设置 `GITHUB_TOKEN` 可以避开 GitHub 搜索接口每分钟 10 次的匿名限额。

## 首屏放什么

首屏最多放两张卡片，左边放趋势榜，右边放主题榜。其余名次都进榜单记录表，并附核验链接。

| 位置 | 显示规则 |
|---|---|
| 左卡 · 趋势榜 | 填了 `trendshift.repositoryId` 时，换成 Trendshift 官方卡片，由 Trendshift 自动更新。没填时从 `trending` 里挑一条名次最好的画成自绘卡片：名次数字越小越优先；名次相同时，月榜优先于周榜，周榜优先于日榜，全语言优先于单一语言，日期新的优先。某一条加上 `"featured": true` 就固定显示它。 |
| 右卡 · 主题榜 | 有 2 个及以上主题排第 1 时，显示「#1 in N GitHub Topics」，点击跳到榜单记录表。只有 1 个主题第 1 时，显示该主题名，点击进入按 Star 排序的主题页。没有主题第 1 时，显示最好的前 3 名。前 3 名都没有就不显示这张卡。 |
| 榜单记录表 | 趋势榜保留最近 10 条，按日期从新到旧排；主题榜列出前 3 名，第 4 名以后只留在 JSON 里。 |

卡片有亮色、暗色两套，README 用 `<picture>` 按读者的 GitHub 主题切换。

## 数据从哪来

| 榜单 | 怎么更新 | 说明 |
|---|---|---|
| GitHub 主题榜（按 Star） | `refresh` 自动更新 | 会先同步仓库当前设置的 topics。新加的主题，脚本会提示补 `name.zh`。 |
| GitHub Trending 日榜 / 周榜 / 月榜 | `refresh` 自动记录 | 检查全语言榜，以及 `trendingLanguages` 里的语言（目前是 TypeScript）。只有 Luna 在榜时才记一条；同一天多次运行，只保留最好的名次。 |
| Trendshift 日榜 / 周榜 | 手工登记 | Trendshift 会拦自动请求。在 Trendshift 页面确认名次后，在 `trending` 里加一条，见下方示例。 |
| Trendshift 官方卡片 | 一次性设置 | 在 Trendshift 打开 Luna 的仓库页，把地址 `trendshift.io/repositories/<编号>` 里的数字填进 `trendshift.repositoryId`。 |

手工登记 Trendshift 名次：

```json
{
  "date": "2026-12-01",
  "source": "trendshift",
  "period": "weekly",
  "scope": "all",
  "rank": 5,
  "url": "https://trendshift.io/repositories/<编号>",
  "snapshot": "2026-12-01 08:00 UTC"
}
```

- `period` 可填 `daily`、`weekly`、`monthly`。
- `scope` 填 `all` 表示全语言，也可以填语言名，比如 `TypeScript`。
- `url` 尽量填能看到 Luna 名次的页面。

## 更新节奏

1. **冲榜期间每天一次**：运行 `refresh`。如果 Trendshift 当天有新名次，手工补一条，然后提交 `docs/rankings.json`、`docs/assets/badges/` 和两份 README。
2. **每次发社区小版本前**：跑一次 `refresh` 再发版，跟版本同步流程放在一起。
3. **拿到 Trendshift 编号后**：填 `repositoryId` 并重新生成，左卡会换成官方卡片。自绘的趋势卡不再使用，但历史名次仍保留在记录表里。
4. **新增主题标签后**：在 GitHub 仓库设置里加好 topic，再跑 `refresh`，然后补中文名。

以后可以把第 1 步做成 GitHub Actions 定时任务，让它每天跑 `refresh` 并开 PR。但那样需要给工作流 `contents: write` 和 `pull-requests: write` 权限，所以暂时没有开启。
