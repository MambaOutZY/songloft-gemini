-- +goose Up
-- 视频支持（songloft-org/songloft#76）：
-- 1) 扩展扫描格式，纳入常见视频容器（mkv/webm/avi/ts），使本地视频文件能被扫描入库；
--    这类文件以"听"为主（网课、电子书、音视频混合内容），至少播放其音频轨，桌面/Web 可渲染画面。
-- 2) songs 增加 is_video 列，标记一首歌是否含真实视频轨（扫描时由 ffprobe 探测，排除封面 attached_pic）。
--    播放页据此决定是否渲染画面；DLNA 投屏据此选择 VideoMime/AudioMime。

-- mkv：仅当数组中尚无 "mkv" 时幂等追加，其余同理。
UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'mkv')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"mkv"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'webm')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"webm"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'avi')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"avi"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'ts')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"ts"%';

-- +goose StatementBegin
ALTER TABLE songs ADD COLUMN is_video INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE songs DROP COLUMN is_video;
-- +goose StatementEnd

-- 从 scan_config.supported_formats 移除视频容器。
UPDATE configs
SET value = json_set(value, '$.supported_formats',
  (SELECT json_group_array(e.value)
     FROM json_each(json_extract(configs.value, '$.supported_formats')) e
    WHERE e.value NOT IN ('mkv', 'webm', 'avi', 'ts')))
WHERE key = 'scan_config'
  AND (value LIKE '%"mkv"%' OR value LIKE '%"webm"%' OR value LIKE '%"avi"%' OR value LIKE '%"ts"%');
