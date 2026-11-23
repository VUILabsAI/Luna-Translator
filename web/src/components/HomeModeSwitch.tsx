import type { WebHomeMode } from '../lib/homeModes'

export function HomeModeSwitch({ value, disabled = false, meetingAvailable, onChange }: {
  value: WebHomeMode
  disabled?: boolean
  meetingAvailable: boolean
  onChange: (mode: WebHomeMode) => void
}) {
  if (value === 'conversation') {
    return <button type="button" className="community-mode-shortcut" disabled={disabled} onClick={() => onChange('solo')}>
      <img src="/showcase/canary-figma/quick-start.png" alt="" />返回快速开始
    </button>
  }

  return <div className="community-mode-shortcuts" role="group" aria-label="同传模式">
    <button type="button" className="community-mode-shortcut" disabled={disabled || !meetingAvailable}
      title={meetingAvailable ? undefined : '社区版暂未提供多人会议'}
      onClick={() => onChange('meeting')}>
      <img src="/showcase/canary-figma/meeting.png" alt="" />会议
    </button>
    <button type="button" className="community-mode-shortcut" disabled={disabled} onClick={() => onChange('conversation')}>
      <img src="/showcase/canary-figma/conversation.png" alt="" />对话
    </button>
  </div>
}
