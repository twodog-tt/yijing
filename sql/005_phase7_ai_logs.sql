-- Phase 7：AI 调用观测日志表

USE yijing;

CREATE TABLE IF NOT EXISTS ai_generation_logs (
  id                BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  divination_id     BIGINT       NOT NULL                COMMENT '关联 divination_record.id',
  question_summary  VARCHAR(80)  NULL                    COMMENT '用户问题摘要，最多80字',
  ai_provider       VARCHAR(32)  NOT NULL                COMMENT 'AI提供方，如 mock/deepseek/mock_fallback',
  model_name        VARCHAR(64)  NOT NULL                COMMENT '模型名称',
  status            INT          NOT NULL                COMMENT '状态：1=成功 2=失败 3=fallback成功',
  duration_ms       INT          NOT NULL DEFAULT 0      COMMENT '调用耗时毫秒',
  fallback_used     INT          NOT NULL DEFAULT 0      COMMENT '是否发生降级：1=是 0=否',
  error_message     VARCHAR(500) NULL                    COMMENT '错误信息摘要，最多500字',
  created_at        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_ai_logs_divination_id (divination_id),
  KEY idx_ai_logs_status (status),
  KEY idx_ai_logs_provider (ai_provider),
  KEY idx_ai_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI完整解读生成调用日志表';
