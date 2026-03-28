-- 真人面试安全闭环迁移脚本
-- 目标：补齐 UUID、邀请码、角色与状态字段，并回填历史数据
-- 适用：MySQL 8+

START TRANSACTION;

-- 1) 用户 UUID
ALTER TABLE users
  ADD COLUMN uuid CHAR(36) NULL AFTER id;

UPDATE users
SET uuid = UUID()
WHERE uuid IS NULL OR uuid = '';

ALTER TABLE users
  ADD UNIQUE INDEX uk_users_uuid (uuid);

-- 2) 面试主表安全字段
ALTER TABLE interviews
  ADD COLUMN invitation_code VARCHAR(64) NULL AFTER interview_mode,
  ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'candidate' AFTER invitation_code;

ALTER TABLE interviews
  MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';

ALTER TABLE interviews
  ADD UNIQUE INDEX uk_interviews_invitation_code (invitation_code);

-- 3) 真人邀请表补齐邀请码与参与者 UUID
ALTER TABLE human_interview_invitations
  ADD COLUMN invitation_code VARCHAR(64) NULL AFTER id,
  ADD COLUMN student_uuid CHAR(36) NULL AFTER student_id,
  ADD COLUMN invitee_uuid CHAR(36) NULL AFTER invitee_user_id;

UPDATE human_interview_invitations
SET invitation_code = UPPER(REPLACE(UUID(), '-', ''))
WHERE invitation_code IS NULL OR invitation_code = '';

UPDATE human_interview_invitations hi
JOIN users su ON su.id = hi.student_id
JOIN users iu ON iu.id = hi.invitee_user_id
SET hi.student_uuid = su.uuid,
    hi.invitee_uuid = iu.uuid
WHERE hi.student_uuid IS NULL OR hi.student_uuid = '' OR hi.invitee_uuid IS NULL OR hi.invitee_uuid = '';

ALTER TABLE human_interview_invitations
  MODIFY COLUMN invitation_code VARCHAR(64) NOT NULL;

ALTER TABLE human_interview_invitations
  ADD UNIQUE INDEX uk_human_interview_invitation_code (invitation_code);

-- 4) 关联回填：将邀请邀请码同步到真人面试记录
UPDATE interviews i
JOIN human_interview_invitations hi ON hi.interview_id = i.id
SET i.invitation_code = hi.invitation_code
WHERE i.interview_mode = 'human'
  AND (i.invitation_code IS NULL OR i.invitation_code = '');

COMMIT;
