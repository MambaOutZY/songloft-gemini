-- +goose Up
-- 扩展扫描格式，纳入常用音频/视频容器：
-- 音频：opus（现代音频编码，YouTube/播客常见）、m4b（有声书）、oga（OGG 音频变体扩展名）
-- 视频：mpg（MPEG Program Stream）、mpeg（mpg 别名扩展名）、m4v（Apple 视频）、
--       flv（Flash 视频）、wmv（Windows Media Video）、rm/rmvb（RealMedia）、3gp（手机录像）
-- 幂等：仅当数组中尚无该格式时追加。

-- 音频格式
UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'opus')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"opus"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'm4b')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"m4b"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'oga')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"oga"%';

-- 视频格式
UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'mpg')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"mpg"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'mpeg')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"mpeg"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'm4v')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"m4v"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'flv')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"flv"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'wmv')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"wmv"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'rm')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"rm"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', 'rmvb')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"rmvb"%';

UPDATE configs
SET value = json_set(value, '$.supported_formats[#]', '3gp')
WHERE key = 'scan_config'
  AND value LIKE '%"supported_formats"%'
  AND value NOT LIKE '%"3gp"%';

-- +goose Down
UPDATE configs
SET value = json_set(value, '$.supported_formats',
  (SELECT json_group_array(e.value)
     FROM json_each(json_extract(configs.value, '$.supported_formats')) e
    WHERE e.value NOT IN ('opus', 'm4b', 'oga', 'mpg', 'mpeg', 'm4v', 'flv', 'wmv', 'rm', 'rmvb', '3gp')))
WHERE key = 'scan_config'
  AND (value LIKE '%"opus"%' OR value LIKE '%"m4b"%' OR value LIKE '%"oga"%'
    OR value LIKE '%"mpg"%' OR value LIKE '%"mpeg"%' OR value LIKE '%"m4v"%'
    OR value LIKE '%"flv"%' OR value LIKE '%"wmv"%' OR value LIKE '%"rm"%'
    OR value LIKE '%"rmvb"%' OR value LIKE '%"3gp"%');
