# 榜单维护

README 首屏的两张榜单卡片和「项目状态 › 榜单记录」表格都由 [`scripts/readme-badges.mjs`](../scripts/readme-badges.mjs) 根据 [`rankings.json`](rankings.json) 生成。不要直接改 README 里 `<!-- rankings:… -->` 标记之间的内容，下次生成会被覆盖。

```sh
node scripts/readme-badges.mjs           # 按 rankings.json 重新生成卡片、按钮和两份 README
node scripts/readme-badges.mjs refresh   # 先查 GitHub 主题榜和 GitHub Trending，再生成
node scripts/readme-badges.mjs import /path/to/readme-rankings.json  # 合并监控导出，再生成
node scripts/readme-badges.mjs --check   # 只检查是否需要重新生成（CI 会跑）
```

`refresh` 读取公开榜页和 GitHub 仓库元数据。匿名额度用完时，可以为元数据请求配置 `GITHUB_TOKEN`。

## 首屏放什么

首屏最多放两张卡片，左边放趋势榜，右边放主题榜。其余名次都进榜单记录表，并附核验链接。

| 位置 | 显示规则 |
|---|---|
| 左卡 · 趋势榜 | 填了 `trendshift.repositoryId` 时，换成 Trendshift 官方卡片，由 Trendshift 自动更新。没填时从 `trending` 里挑一条名次最好的画成自绘卡片：名次数字越小越优先；名次相同时，月榜优先于周榜，周榜优先于日榜，全语言优先于单一语言，日期新的优先。某一条加上 `"featured": true` 就固定显示它。 |
| 右卡 · 主题榜 | 有 2 个及以上主题排第 1 时，显示「#1 in N GitHub Topics」，点击跳到榜单记录表。只有 1 个主题第 1 时，显示该主题名，点击进入按 Star 排序的主题页。没有主题第 1 时，显示最好的前 3 名。前 3 名都没有就不显示这张卡。 |
| 榜单记录表 | 首次核验成绩和每次刷新历史最佳各占一行，显示名次、获得时间、可点击徽章和说明。依次按名次、全站榜优先、获得时间从新到旧排序。全站榜指全语言趋势榜，各语言榜和 GitHub 主题榜分别统计。 |

榜单记录表的徽章由脚本生成在 `docs/assets/badges/rankings/`，点击进入对应榜页。说明列标记该榜单的最新历史最佳，旧记录继续保留。

卡片有亮色、暗色两套，README 用 `<picture>` 按读者的 GitHub 主题切换。

## 数据从哪来

| 榜单 | 怎么更新 | 说明 |
|---|---|---|
| GitHub 主题榜（按 Star） | 本地监控导入或 `refresh` | 使用主题页按 Star 排序后的结果。新加的主题可以补上 `name.zh`。 |
| GitHub Trending 日榜 / 周榜 | 本地监控导入或 `refresh` | `refresh` 检查全语言和 `trendingLanguages` 配置中的语言，也支持月榜。 |
| Trendshift 日榜 / 周榜 | 本地监控导入或手工登记 | 从监控导入公开榜页的核验结果。手工登记时在 `recordHistory` 添加成绩。首屏趋势卡的候选成绩由 `trending` 提供。 |
| Trendshift 官方卡片 | 一次性设置 | 在 Trendshift 打开 Luna 的仓库页，把地址 `trendshift.io/repositories/<编号>` 里的数字填进 `trendshift.repositoryId`。 |

`recordHistory` 中的条目示例

```json
{
  "date": "2026-11-26",
  "source": "trendshift",
  "period": "daily",
  "scope": "TypeScript",
  "rank": 3,
  "url": "https://trendshift.io/?language=TypeScript",
  "observedAt": "2026-11-26T06:23:50Z"
}
```

- 趋势榜的 `source` 为 `github` 或 `trendshift`，`period` 可为 `daily`、`weekly` 或 `monthly`。全语言用 `all`，语言榜使用语言名。
- 主题榜用 `source: "topic"`、`period: "stars"` 和主题标签作为 `scope`。通过 `url` 链接到按 Star 排序的主题页。
- `date` 和 `observedAt` 均按 UTC 填写。旧数据只核验到日期时，保留 `date` 即可。一天内多次刷新时，用 `observedAt` 保留先后顺序。

## 更新节奏

1. 检查监控记录中的榜单链接和核验时间，导入最新的 `readme-rankings.json`。导入会合并已有成绩，重复导入同一份数据不会增加条目。
2. 运行 `node scripts/readme-badges.mjs --check`，再检查 `docs/rankings.json`、两份 README 和随排名更新的卡片。
