-- Phase 3：补充 interpretation 状态说明（不破坏已有结构）

USE yijing;

ALTER TABLE interpretation
  MODIFY COLUMN generation_status INT NOT NULL DEFAULT 0
    COMMENT '生成状态：0=待生成 1=免费已生成 2=完整已生成 3=生成失败';
