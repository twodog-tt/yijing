-- 事项类型种子数据（第一版固定 5 类，无综合/其他）

USE yijing;

INSERT INTO matter_category (id, code, name, description, sort_order, status)
VALUES
  (1, 'career',        '事业',     '工作、职业发展与职场方向相关问事',       1, 1),
  (2, 'relationship',  '关系',     '感情、人际与家庭关系相关问事',           2, 1),
  (3, 'study',         '学习',     '学业、技能提升与认知成长相关问事',       3, 1),
  (4, 'choice',        '选择',     '面临抉择时的方向参考与自我反思',         4, 1),
  (5, 'recent_state',  '近期状态', '近期心态、状态与生活节奏相关问事',       5, 1)

ON DUPLICATE KEY UPDATE
  code = VALUES(code),
  name = VALUES(name),
  description = VALUES(description),
  sort_order = VALUES(sort_order),
  status = VALUES(status),
  updated_at = CURRENT_TIMESTAMP;

-- 高风险敏感词（硬拦截）
INSERT INTO sensitive_keyword (keyword, category, block_message, status)
VALUES
  ('癌症',     'medical',    '本工具不提供医疗诊断或疾病预测，请勿就医疗问题起卦。', 1),
  ('肿瘤',     'medical',    '本工具不提供医疗诊断或疾病预测，请勿就医疗问题起卦。', 1),
  ('治病',     'medical',    '本工具不提供医疗诊断或疾病预测，请勿就医疗问题起卦。', 1),
  ('吃药',     'medical',    '本工具不提供医疗诊断或疾病预测，请勿就医疗问题起卦。', 1),
  ('手术',     'medical',    '本工具不提供医疗诊断或疾病预测，请勿就医疗问题起卦。', 1),
  ('寿命',     'death',      '本工具不涉及寿命与生死预测，请勿就此类问题起卦。', 1),
  ('还能活',   'death',      '本工具不涉及寿命与生死预测，请勿就此类问题起卦。', 1),
  ('自杀',     'self_harm',  '检测到自伤相关内容，请寻求专业帮助，本工具无法提供此类解读。', 1),
  ('自残',     'self_harm',  '检测到自伤相关内容，请寻求专业帮助，本工具无法提供此类解读。', 1),
  ('跳楼',     'self_harm',  '检测到自伤相关内容，请寻求专业帮助，本工具无法提供此类解读。', 1),
  ('股票',     'investment', '本工具不提供投资涨跌预测，请勿就投资问题起卦。', 1),
  ('涨停',     'investment', '本工具不提供投资涨跌预测，请勿就投资问题起卦。', 1),
  ('基金',     'investment', '本工具不提供投资涨跌预测，请勿就投资问题起卦。', 1),
  ('比特币',   'investment', '本工具不提供投资涨跌预测，请勿就投资问题起卦。', 1),
  ('彩票',     'gambling',   '本工具不涉及赌博与彩票预测，请勿就此类问题起卦。', 1),
  ('博彩',     'gambling',   '本工具不涉及赌博与彩票预测，请勿就此类问题起卦。', 1),
  ('赌场',     'gambling',   '本工具不涉及赌博与彩票预测，请勿就此类问题起卦。', 1),
  ('起诉',     'legal',      '本工具不提供法律诉讼建议，请咨询专业律师。', 1),
  ('官司',     'legal',      '本工具不提供法律诉讼建议，请咨询专业律师。', 1),
  ('坐牢',     'legal',      '本工具不提供法律诉讼建议，请咨询专业律师。', 1),
  ('违法',     'illegal',    '本工具不支持违法行为相关内容，请勿起卦。', 1),
  ('犯罪',     'illegal',    '本工具不支持违法行为相关内容，请勿起卦。', 1),
  ('贩毒',     'illegal',    '本工具不支持违法行为相关内容，请勿起卦。', 1),
  ('报复',     'threat',     '本工具不支持恐吓报复相关内容，请勿起卦。', 1),
  ('杀人',     'threat',     '本工具不支持恐吓报复相关内容，请勿起卦。', 1)

ON DUPLICATE KEY UPDATE
  category = VALUES(category),
  block_message = VALUES(block_message),
  status = VALUES(status);
