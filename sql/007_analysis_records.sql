-- Phase D：通用分析报告表（八字 / 奇门）
-- 约束：不使用 ENUM / TINYINT / CHECK；状态字段使用 INT
-- 仅新增 analysis_records，不修改现有表

USE yijing;

CREATE TABLE IF NOT EXISTS analysis_records (
  id                 BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  session_id         BIGINT      NOT NULL COMMENT '关联 user_sessions.id',
  module_type        INT         NOT NULL COMMENT '模块类型：1=八字 2=奇门',
  algorithm_version  VARCHAR(32) NOT NULL COMMENT '分析规则版本，如 bazi-simple-v1',
  category_id        BIGINT      NULL COMMENT '事项类型ID，部分模块可为空',
  question           VARCHAR(500) NULL COMMENT '用户问题，八字模块为空',
  input_payload      JSON        NOT NULL COMMENT '服务端生成的输入快照JSON',
  result_payload     JSON        NULL COMMENT '非AI计算产生的结构化结果JSON',
  free_content       TEXT        NULL COMMENT '免费解读正文',
  full_content       MEDIUMTEXT  NULL COMMENT '完整解读正文',
  unlock_status      INT         NOT NULL DEFAULT 0 COMMENT '解锁状态：0=未解锁 1=已解锁',
  unlock_type        VARCHAR(32) NULL COMMENT '成功解锁方式',
  ai_provider        VARCHAR(32) NULL COMMENT '完整解读实际使用的AI提供方',
  generation_status  INT         NOT NULL DEFAULT 0 COMMENT '生成状态：0=待生成 1=免费完成 2=完整生成中 3=完整完成 4=免费失败 5=完整失败',
  status             INT         NOT NULL DEFAULT 1 COMMENT '记录状态：1=正常 0=删除',
  created_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_analysis_session_status_created (session_id, status, created_at, id),
  KEY idx_analysis_session_module_status_created (session_id, module_type, status, created_at, id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='通用分析报告表';
