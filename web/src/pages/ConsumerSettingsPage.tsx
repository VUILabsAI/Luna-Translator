import { openAIConsentSettings } from '../lib/aiConsent'
import { logout } from '@simultrans/ui-shared/src/lib/consumerSession'
import ConsumerPageFrame from '../components/ConsumerPageFrame'

export default function ConsumerSettingsPage() {
  return <ConsumerPageFrame className="community-preferences" labelledBy="community-preferences-title" backHref="/">
    <h1 id="community-preferences-title">社区版设置</h1>
    <section className="community-preferences-group" aria-labelledby="community-workspace-title">
      <h2 id="community-workspace-title">工作区</h2>
      <p>同一部署密码的使用者共享会话历史和模型额度。不同用户可创建独立部署。</p>
      <a href="/history">历史记录 <span aria-hidden="true">›</span></a>
    </section>
    <section className="community-preferences-group" aria-labelledby="community-data-title">
      <h2 id="community-data-title">模型与数据</h2>
      <p>通过后端环境变量配置模型服务。会话记录和音频存储在 Docker 数据卷。分享链接持有者可查看相应记录及音频。</p>
      <button type="button" onClick={openAIConsentSettings}>AI 数据授权 <span aria-hidden="true">›</span></button>
    </section>
    <button type="button" className="community-sign-out" onClick={() => void logout().finally(() => window.location.assign('/'))}>退出工作区</button>
  </ConsumerPageFrame>
}
