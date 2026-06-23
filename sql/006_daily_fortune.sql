-- Phase 9：今日运势

USE yijing;

INSERT INTO matter_category (id, code, name, description, sort_order, status)
VALUES
  (6, 'daily_fortune', '今日运势', '今日整体状态、节奏与行动提醒', 6, 1)
ON DUPLICATE KEY UPDATE
  code = VALUES(code),
  name = VALUES(name),
  description = VALUES(description),
  sort_order = VALUES(sort_order),
  status = VALUES(status),
  updated_at = CURRENT_TIMESTAMP;

CREATE TABLE IF NOT EXISTS daily_fortunes (
  id             BIGINT   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  session_id     BIGINT   NOT NULL                COMMENT '关联 user_sessions.id',
  fortune_date   DATE     NOT NULL                COMMENT '运势日期（用户本地日期）',
  divination_id  BIGINT   NOT NULL                COMMENT '关联 divination_record.id',
  status         INT      NOT NULL DEFAULT 1      COMMENT '状态：1=正常 0=删除',
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_session_date (session_id, fortune_date),
  KEY idx_divination_id (divination_id),
  KEY idx_fortune_date (fortune_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='每日运势映射表';
