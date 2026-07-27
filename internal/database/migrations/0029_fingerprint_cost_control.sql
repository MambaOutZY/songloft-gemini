-- +goose Up
-- 指纹计算「已尝试」标记（unix 秒，0 = 未尝试）。
-- 此前失败（超时 / 无音轨 / 损坏文件）不落库，每轮扫描都会把同一批注定失败的文件
-- 重新捞出来跑 ffmpeg 全解码，形成永久 CPU 占用（songloft-org/songloft#323）。
-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN fingerprint_attempted_at INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- 指纹采样改为 AcoustID 事实标准的前 120 秒，与旧的「整文件全长」指纹不可比较，
-- 一次性清空，由用户在「重复歌曲检测」页或开启自动指纹后重算。
UPDATE songs SET fingerprint = '', fingerprint_duration = 0 WHERE fingerprint != '';

-- 扫描后自动计算指纹，默认关闭（指纹只服务于重复检测与插件搜索，属按需功能）。
INSERT OR IGNORE INTO configs (key, value) VALUES ('scan_auto_fingerprint', 'false');

-- +goose Down
DELETE FROM configs WHERE key = 'scan_auto_fingerprint';
