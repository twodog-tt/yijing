-- AI 易经问事 MVP - 数据库结构
-- 约束：不使用 ENUM / TINYINT / CHECK；状态字段使用 INT

CREATE DATABASE IF NOT EXISTS yijing
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE yijing;

-- ---------------------------------------------------------------------------
-- 匿名用户会话表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_sessions (
  id            BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  session_key   VARCHAR(64)  NOT NULL                COMMENT '会话唯一标识，前端生成或后端分配',
  client_info   VARCHAR(255) NULL                    COMMENT '客户端信息，如 User-Agent',
  status        INT          NOT NULL DEFAULT 1      COMMENT '状态：1=有效 0=失效',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_sessions_session_key (session_key),
  KEY idx_user_sessions_status (status),
  KEY idx_user_sessions_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='匿名用户会话表';

-- ---------------------------------------------------------------------------
-- 事项类型表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS matter_category (
  id            BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  code          VARCHAR(32)  NOT NULL                COMMENT '类型编码，如 career',
  name          VARCHAR(64)  NOT NULL                COMMENT '类型显示名称',
  description   VARCHAR(255) NULL                    COMMENT '类型说明',
  sort_order    INT          NOT NULL DEFAULT 0      COMMENT '排序权重，越小越靠前',
  status        INT          NOT NULL DEFAULT 1      COMMENT '状态：1=启用 0=禁用',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_matter_category_code (code),
  KEY idx_matter_category_status_sort (status, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='问事事项类型表';

-- ---------------------------------------------------------------------------
-- 六十四卦基础数据表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hexagram (
  id              BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  number          INT          NOT NULL                COMMENT '卦序，1-64',
  name            VARCHAR(16)  NOT NULL                COMMENT '卦名，如乾',
  full_name       VARCHAR(32)  NOT NULL                COMMENT '全名，如乾为天',
  upper_trigram   VARCHAR(8)   NOT NULL                COMMENT '上卦名',
  lower_trigram   VARCHAR(8)   NOT NULL                COMMENT '下卦名',
  binary_code     VARCHAR(6)   NOT NULL                COMMENT '六位二进制，自下而上，1=阳爻 0=阴爻',
  judgment        TEXT         NULL                    COMMENT '卦辞',
  image_text      TEXT         NULL                    COMMENT '象辞',
  summary         VARCHAR(512) NULL                    COMMENT '一句话摘要',
  status          INT          NOT NULL DEFAULT 1      COMMENT '状态：1=启用 0=禁用',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_hexagram_number (number),
  UNIQUE KEY uk_hexagram_binary_code (binary_code),
  KEY idx_hexagram_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='六十四卦基础数据表';

-- ---------------------------------------------------------------------------
-- 问事起卦记录表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS divination_record (
  id                    BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  session_id            BIGINT       NOT NULL                COMMENT '关联 user_sessions.id',
  category_id           BIGINT       NOT NULL                COMMENT '关联 matter_category.id',
  question              VARCHAR(500) NOT NULL                COMMENT '用户输入的问题',
  method                VARCHAR(32)  NOT NULL DEFAULT 'coin_three' COMMENT '起卦方法，如 coin_three',
  primary_hexagram_id   BIGINT       NOT NULL                COMMENT '本卦ID，关联 hexagram.id',
  changed_hexagram_id   BIGINT       NULL                    COMMENT '变卦ID，无动爻时可为空',
  moving_lines          VARCHAR(32)  NOT NULL DEFAULT '[]'    COMMENT '动爻位置JSON数组字符串，如 [2,5]',
  line_snapshot         TEXT         NOT NULL                COMMENT '六爻快照JSON字符串',
  seed                  VARCHAR(64)  NOT NULL                COMMENT '起卦种子，用于结果追溯',
  unlock_status         INT          NOT NULL DEFAULT 0      COMMENT '解锁状态：0=未解锁 1=已解锁',
  status                INT          NOT NULL DEFAULT 1      COMMENT '记录状态：1=正常 0=删除',
  created_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_divination_session_created (session_id, created_at),
  KEY idx_divination_category_id (category_id),
  KEY idx_divination_primary_hexagram (primary_hexagram_id),
  KEY idx_divination_unlock_status (unlock_status),
  KEY idx_divination_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='问事起卦记录表';

-- ---------------------------------------------------------------------------
-- 六爻明细表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS divination_line (
  id              BIGINT   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  divination_id   BIGINT   NOT NULL                COMMENT '关联 divination_record.id',
  line_position   INT      NOT NULL                COMMENT '爻位，1-6，自下而上',
  line_value      INT      NOT NULL                COMMENT '爻值，6=老阴 7=少阳 8=少阴 9=老阳',
  is_yang         INT      NOT NULL                COMMENT '是否阳爻：1=阳 0=阴',
  is_moving       INT      NOT NULL                COMMENT '是否动爻：1=是 0=否',
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_divination_line_position (divination_id, line_position),
  KEY idx_divination_line_divination_id (divination_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='起卦六爻明细表';

-- ---------------------------------------------------------------------------
-- 解读内容表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS interpretation (
  id                  BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  divination_id       BIGINT       NOT NULL                COMMENT '关联 divination_record.id，一问一解读',
  free_content        TEXT         NOT NULL                COMMENT '免费解读内容',
  full_content        TEXT         NULL                    COMMENT '完整解读内容，解锁后生成',
  ai_provider         VARCHAR(32)  NOT NULL DEFAULT 'mock' COMMENT 'AI提供方，如 mock / deepseek',
  generation_status   INT          NOT NULL DEFAULT 0      COMMENT '生成状态：0=待生成 1=免费已生成 2=完整已生成',
  generated_at        DATETIME     NULL                    COMMENT '解读生成完成时间',
  created_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_interpretation_divination_id (divination_id),
  KEY idx_interpretation_generation_status (generation_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='卦象解读内容表';

-- ---------------------------------------------------------------------------
-- 解锁记录表（mock 广告/支付）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS unlock_record (
  id                    BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  divination_id         BIGINT       NOT NULL                COMMENT '关联 divination_record.id',
  session_id            BIGINT       NOT NULL                COMMENT '关联 user_sessions.id',
  unlock_type           VARCHAR(32)  NOT NULL                COMMENT '解锁方式，如 mock_ad / mock_button',
  unlock_status         INT          NOT NULL DEFAULT 1      COMMENT '解锁结果：1=成功 0=失败',
  mock_transaction_id   VARCHAR(64)  NOT NULL                COMMENT '模拟交易流水号',
  created_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_unlock_divination_id (divination_id),
  KEY idx_unlock_session_id (session_id),
  KEY idx_unlock_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='完整解读解锁记录表';

-- ---------------------------------------------------------------------------
-- 高风险敏感词表（硬拦截）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sensitive_keyword (
  id                BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  keyword           VARCHAR(64)  NOT NULL                COMMENT '敏感关键词',
  category          VARCHAR(32)  NOT NULL                COMMENT '分类：medical/death/self_harm/investment/gambling/legal/illegal/threat',
  block_message     VARCHAR(255) NOT NULL                COMMENT '拦截提示文案',
  status            INT          NOT NULL DEFAULT 1      COMMENT '状态：1=启用 0=禁用',
  created_at        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_sensitive_keyword (keyword),
  KEY idx_sensitive_keyword_category (category),
  KEY idx_sensitive_keyword_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='高风险问题敏感词拦截表';
