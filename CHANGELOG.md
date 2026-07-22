## [v2.11.0] - 2026-07-22
### :sparkles: New Features
- [`7dc4d92`](https://github.com/songloft-org/songloft/commit/7dc4d9288f024440da761bf9cca73a446a6894cc) - **jsplugin**: 为 JS 插件 SDK 增加 TCP Socket API *(PR [#276](https://github.com/songloft-org/songloft/pull/276) by [@hanxi](https://github.com/hanxi))*
- [`8687a3e`](https://github.com/songloft-org/songloft/commit/8687a3e9d87a1ba810ecf220e48fe94e28de508a) - **jsplugin**: 插件自动更新开关 + 源列表「全部」聚合 *(PR [#270](https://github.com/songloft-org/songloft/pull/270) by [@hanxi](https://github.com/hanxi))*
- [`09c292a`](https://github.com/songloft-org/songloft/commit/09c292abb21d9799e5b9b89b02eafe59b99ff466) - **songs**: 按流派/语种/风格等标签分类浏览曲库 *(PR [#277](https://github.com/songloft-org/songloft/pull/277) by [@hanxi](https://github.com/hanxi))*
- [`548bee6`](https://github.com/songloft-org/songloft/commit/548bee6769055b9710e7c275b905bdbbb4b62630) - **scan**: 支持 mp4 格式音频扫描与播放 *(PR [#281](https://github.com/songloft-org/songloft/pull/281) by [@hanxi](https://github.com/hanxi))*
- [`338d574`](https://github.com/songloft-org/songloft/commit/338d574dcfdfeebbaa2c23aa19d0aa592fabeccb) - **scan**: 支持视频容器扫描、播放与 DLNA 视频投屏 *(PR [#76](https://github.com/songloft-org/songloft/pull/76) by [@hanxi](https://github.com/hanxi))*
- [`1aa9936`](https://github.com/songloft-org/songloft/commit/1aa99368e13413f6a76922fc479e5907b4d31cb7) - **songs**: 网络歌曲/电台创建与更新端点支持 is_video *(PR [#76](https://github.com/songloft-org/songloft/pull/76) by [@hanxi](https://github.com/hanxi))*
- [`89b788e`](https://github.com/songloft-org/songloft/commit/89b788e52f4667ef0a6fca64cafcbc7353608d84) - **playlist**: ListPlaylists 支持 keyword 搜索,并同步前端子模块 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a6d7441`](https://github.com/songloft-org/songloft/commit/a6d744106a9439b569eec1a1c7b5789079def004) - **songs**: facets 支持分页/搜索/封面 + 新增 /settings/library-browse 配置端点 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a650185`](https://github.com/songloft-org/songloft/commit/a6501858343cfdce239a6a2f2cccfa0c9fc82835) - **songs**: 下载支持可选转码(format/quality),复用 GetOrTranscode hanxi/songloft-plugin-bili[#1](https://github.com/songloft-org/songloft/pull/1) *(commit by [@hanxi](https://github.com/hanxi))*
- [`692ded7`](https://github.com/songloft-org/songloft/commit/692ded7acf97e71f75bf2c14a6f9a06ac1311b27) - **songs**: library-browse 放行歌单三视图 key，默认顺序按组连续 *(commit by [@hanxi](https://github.com/hanxi))*
- [`edf6d99`](https://github.com/songloft-org/songloft/commit/edf6d99a2f38da70102877aadb98e2c01a795b09) - **jsplugin**: 客户端 SDK 宿主桥接(common.js) + 文档 *(PR [#285](https://github.com/songloft-org/songloft/pull/285) by [@hanxi](https://github.com/hanxi))*
- [`441f1a6`](https://github.com/songloft-org/songloft/commit/441f1a6b136a2ffbc7651b456acf60a2e6938117) - 支持 MKA 导入/播放与双音轨切换 *(commit by [@hanxi](https://github.com/hanxi))*
- [`c3b62e3`](https://github.com/songloft-org/songloft/commit/c3b62e32cd9e56a6044a5cd3bf413a696df1a06a) - **lyric**: /songs/{id}/lyric 支持 refresh 强制重抓歌词 *(PR [#303](https://github.com/songloft-org/songloft/pull/303) by [@hanxi](https://github.com/hanxi))*
- [`d219948`](https://github.com/songloft-org/songloft/commit/d219948a0e8dc27918ac94a618b4522bbd93a505) - CUE 按需提取替代预分割，解决磁盘空间占用问题 *(PR [#306](https://github.com/songloft-org/songloft/pull/306) by [@hanxi](https://github.com/hanxi))*
- [`b964bbf`](https://github.com/songloft-org/songloft/commit/b964bbf1f76239614699c21e1208894a61d987b8) - **cache**: 缓存网络歌曲支持统一转码落盘格式 [#300](https://github.com/songloft-org/songloft/pull/300) *(commit by [@hanxi](https://github.com/hanxi))*
- [`9da8d62`](https://github.com/songloft-org/songloft/commit/9da8d628641c726749d2696d6a5a0ffba0d2609e) - **logs**: 新增日志落盘与脱敏导出端点 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a221545`](https://github.com/songloft-org/songloft/commit/a22154560438b8258f744412ded29b9499d88b7b) - **radio**: 电台流支持服务端实时转码，兼容仅支持 MP3 的播放设备 [#275](https://github.com/songloft-org/songloft/pull/275) *(commit by [@hanxi](https://github.com/hanxi))*
- [`88e9abb`](https://github.com/songloft-org/songloft/commit/88e9abb23283b6b07cbec70fc9e315d71935b5ea) - 插件sdk新增封面提供者接口 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`878d9c0`](https://github.com/songloft-org/songloft/commit/878d9c0dca98510d8b3f973e492a377b3583bcd6) - **radio**: 电台代理改为 ICY 元数据透传，修复 web 端播放 1 秒即停 *(PR [#275](https://github.com/songloft-org/songloft/pull/275) by [@hanxi](https://github.com/hanxi))*
- [`387058e`](https://github.com/songloft-org/songloft/commit/387058ed6c04290de3c9d8f0089286341c28d64d) - **download**: 缓解导入探测与批量下载争用导致的下载失败 *(PR [#265](https://github.com/songloft-org/songloft/pull/265) by [@hanxi](https://github.com/hanxi))*
- [`a6509fa`](https://github.com/songloft-org/songloft/commit/a6509fa0332ad066ae242885685d76afdf6056a5) - **download**: 用停滞检测替代整请求硬超时，避免慢速网络下载被误掐 *(PR [#265](https://github.com/songloft-org/songloft/pull/265) by [@hanxi](https://github.com/hanxi))*
- [`2c14108`](https://github.com/songloft-org/songloft/commit/2c1410895aef25127260ad52c8ac7ed6e2563431) - **jsplugin**: 自动更新用 GetJSON 读 github_proxy 配置,修正代理前缀拼接 *(commit by [@hanxi](https://github.com/hanxi))*
- [`91a1e44`](https://github.com/songloft-org/songloft/commit/91a1e44b36f8e7f41113e48cca62346e1bcf2321) - **web**: app shell 静态资源改用 no-cache,修复升级后浏览器仍跑旧 main.dart.js *(commit by [@hanxi](https://github.com/hanxi))*
- [`c63531f`](https://github.com/songloft-org/songloft/commit/c63531f2379347d2dc758720ae2742172d004102) - **player**: 更新子模块修复 Web 插件 Tab iframe 反复重载抖动 *(PR [#278](https://github.com/songloft-org/songloft/pull/278) by [@hanxi](https://github.com/hanxi))*
- [`e40b702`](https://github.com/songloft-org/songloft/commit/e40b702ed7c1bf0f28e3e6192f19bc3144c2d966) - **player**: 更新子模块修复移动端插件 Tab 再次打开黑屏/底栏消失 *(PR [#273](https://github.com/songloft-org/songloft/pull/273) by [@hanxi](https://github.com/hanxi))*
- [`5df23a2`](https://github.com/songloft-org/songloft/commit/5df23a2ba29728081b50f757e6420e4471877547) - **radio**: 去交织浏览器路径 ICY 元数据,修复 web 端非 m3u8 电台 2-3 秒断流 *(PR [#275](https://github.com/songloft-org/songloft/pull/275) by [@hanxi](https://github.com/hanxi))*
- [`8ecddd4`](https://github.com/songloft-org/songloft/commit/8ecddd4348d9d8435313a3097bdc71aa63cd0c0b) - **source**: 下载链路加同源重试与 Content-Length 截断校验 *(PR [#265](https://github.com/songloft-org/songloft/pull/265) by [@hanxi](https://github.com/hanxi))*
- [`26fb2e1`](https://github.com/songloft-org/songloft/commit/26fb2e13d96c646c600fb1b0b0fdbc0877f7df88) - **jsruntime**: 修复 youtube 歌单串号/标题被改/缓存不全 songloft-org/songloft[#286](https://github.com/songloft-org/songloft/pull/286) *(commit by [@hanxi](https://github.com/hanxi))*
- [`aa5fc0e`](https://github.com/songloft-org/songloft/commit/aa5fc0e6088bf5a4fcbf0a475171f52fe6ca7f19) - **radio**: 电台代理支持 hls=direct 绕过反代 + 归一化 audio/aacp *(commit by [@hanxi](https://github.com/hanxi))*
- [`14e662c`](https://github.com/songloft-org/songloft/commit/14e662c939843d9cd4a7a5d9ffaa678c7e04dc83) - **jsplugin**: 慢端点支持 X-Plugin-Timeout-Ms 放宽调用超时，修复 extract 504 *(PR [#265](https://github.com/songloft-org/songloft/pull/265) by [@hanxi](https://github.com/hanxi))*
- [`3980309`](https://github.com/songloft-org/songloft/commit/398030937f9a7017b25914f527c966d0ddd9834c) - **play**: 慢音源播放解析被切歌/客户端超时误杀致 502 (songloft-org/songloft[#271](https://github.com/songloft-org/songloft/pull/271)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3b1c97b`](https://github.com/songloft-org/songloft/commit/3b1c97b649d171741492f0b5fcf6a3900efa3907) - **radio**: 直连电台流改用非浏览器 UA，修复防盗链源约 3 秒断流 *(PR [#275](https://github.com/songloft-org/songloft/pull/275) by [@hanxi](https://github.com/hanxi))*
- [`dba7966`](https://github.com/songloft-org/songloft/commit/dba796697e3d52003894a68504184b130e18fb91) - docker cache *(commit by [@hanxi](https://github.com/hanxi))*
- [`2b6f113`](https://github.com/songloft-org/songloft/commit/2b6f113bf8f8ad2433fc16d8714bd048e9fce3db) - flutter web 后台回来黑屏问题修复 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e32844b`](https://github.com/songloft-org/songloft/commit/e32844b5498804dcf40a935d468285fb01400c5f) - **lyric**: 修复歌词插件自动抓取失效（测试成功但播放抓不到）[#303](https://github.com/songloft-org/songloft/pull/303) *(commit by [@hanxi](https://github.com/hanxi))*
- [`0e97a99`](https://github.com/songloft-org/songloft/commit/0e97a993998e5c5ac918e77d8ee36c357b7f855f) - **jsplugin**: 修复插件市场分页顺序随机跳变 [#302](https://github.com/songloft-org/songloft/pull/302) *(commit by [@hanxi](https://github.com/hanxi))*
- [`bf21354`](https://github.com/songloft-org/songloft/commit/bf2135491ebbdd2c562010e75a88294ef3384859) - **jsplugin**: 插件页预留滚动条槽根除内容抖动 [#278](https://github.com/songloft-org/songloft/pull/278) *(commit by [@hanxi](https://github.com/hanxi))*
- [`15fb3f8`](https://github.com/songloft-org/songloft/commit/15fb3f80357096533e2566ae9238012ccca5a354) - **source**: 分块 Range 下载绕过 YouTube 单连接限速 (songloft-org/songloft[#305](https://github.com/songloft-org/songloft/pull/305)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`d5d4102`](https://github.com/songloft-org/songloft/commit/d5d4102995fae64bd1d131388d78fb666078d256) - **source**: 流式代理去掉整请求硬超时，避免音箱播到中途重拉/切歌 (songloft-org/songloft-plugin-miot[#55](https://github.com/songloft-org/songloft/pull/55)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`6f7e528`](https://github.com/songloft-org/songloft/commit/6f7e528f040782da784ce368e97abd11d57c3f33) - **jsplugin**: 插件页 html 常驻纵向滚动条根除抖动 [#278](https://github.com/songloft-org/songloft/pull/278) *(commit by [@hanxi](https://github.com/hanxi))*
- [`1cea176`](https://github.com/songloft-org/songloft/commit/1cea1764e687890fb25d3724bf1dd0199a9e433d) - **playactivity**: Activate 不再取消下一首的 prefetch 转码 [#300](https://github.com/songloft-org/songloft/pull/300) *(commit by [@hanxi](https://github.com/hanxi))*
- [`07cea28`](https://github.com/songloft-org/songloft/commit/07cea28e6e404b4e86ef5d2addd8e876bd13c91c) - **jsplugin**: 公共资源 URL 加内容哈希版本号，修复 immutable 缓存致 common.css 更新不下发 [#278](https://github.com/songloft-org/songloft/pull/278) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3d660d0`](https://github.com/songloft-org/songloft/commit/3d660d0f60413786152892649ca5340eb216b305) - **download**: 歌名含斜杠不再拆目录、目标冲突追加序号防覆盖 [#265](https://github.com/songloft-org/songloft/pull/265) *(commit by [@hanxi](https://github.com/hanxi))*
- [`6f373b2`](https://github.com/songloft-org/songloft/commit/6f373b27786767637252d8894a0cd9beb675bc7c) - 下载歌曲后保留 plugin_entry_path，修复重复导入时唯一约束冲突 *(commit by [@hanxi](https://github.com/hanxi))*
- [`b83464b`](https://github.com/songloft-org/songloft/commit/b83464b33ca8641d3f3eb817055c39239825f7ca) - 客户端超时断开时触发后台缓存，避免未缓存歌曲反复失败 *(commit by [@hanxi](https://github.com/hanxi))*
- [`4446b0c`](https://github.com/songloft-org/songloft/commit/4446b0c4c1ab2acea784a421ccb7faada6efa650) - 修复开发版本更新问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ecce7c5`](https://github.com/songloft-org/songloft/commit/ecce7c5da13011a1cfaf92d555cf9cdc9e5025cb) - 修复无法添加超过3万首歌到歌单的问题 *(PR [#308](https://github.com/songloft-org/songloft/pull/308) by [@hanxi](https://github.com/hanxi))*
- [`cbb00c4`](https://github.com/songloft-org/songloft/commit/cbb00c40385a2da05241d083af850d462af0e448) - 已是 MP3 格式的缓存歌曲在设备上播放失败 (songloft-org/songloft[#300](https://github.com/songloft-org/songloft/pull/300)) *(commit by [@hanxi](https://github.com/hanxi))*

### :zap: Performance Improvements
- [`b5f44e1`](https://github.com/songloft-org/songloft/commit/b5f44e1e50d661f5416883619db9eddbef44afee) - **docker**: 交叉编译替代 QEMU 编译 + 分层重排提速镜像打包与更新 *(commit by [@hanxi](https://github.com/hanxi))*
- [`9b0c733`](https://github.com/songloft-org/songloft/commit/9b0c733ec9b8fbb668d99b7495738c68d4af2a68) - **docker**: pin alpine base 与 ffmpeg 镜像 digest 稳定前置层缓存 *(commit by [@hanxi](https://github.com/hanxi))*

### :recycle: Refactors
- [`34b0f7f`](https://github.com/songloft-org/songloft/commit/34b0f7ffea76f7c7cdd52bff81f1157c456d0272) - **jsplugin**: 清理既有 lint *(commit by [@hanxi](https://github.com/hanxi))*
- [`791812f`](https://github.com/songloft-org/songloft/commit/791812fc4b1322a303bca4b492f7fd23475ba15b) - 修正 prefetch 兜底注释并清理两处 lint [#300](https://github.com/songloft-org/songloft/pull/300) *(commit by [@hanxi](https://github.com/hanxi))*

### :construction_worker: Build System
- [`80d136a`](https://github.com/songloft-org/songloft/commit/80d136a75096e86643f9037204d0452d7df31e0c) - web 构建走 beta 3.47(修复切后台白屏),原生升 stable 3.44.6 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f9ab2bd`](https://github.com/songloft-org/songloft/commit/f9ab2bd77392c358e7ec94e381c76daccaf85529) - web 也统一 stable 3.44.6(撤回 web 专用 beta 3.47) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`21ef81d`](https://github.com/songloft-org/songloft/commit/21ef81d64910872f5a252290e99230969076b6da) - update CHANGELOG for v2.10.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`7b4e6ba`](https://github.com/songloft-org/songloft/commit/7b4e6ba08e60bdbd1c61290878a401e9a698f3bc) - **faq**: 新增电台 HLS 流无法播放时开启 HLS 代理的说明 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e0c4e70`](https://github.com/songloft-org/songloft/commit/e0c4e709a22746f1e576711a2606ee0c6021b4cc) - **faq**: 补充应用内视频画面渲染的平台支持说明 (songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`de95b87`](https://github.com/songloft-org/songloft/commit/de95b87738c31aa85d68a288937e2a429b1f11ac) - **frontend**: 更新音频后端与视频画面架构描述 *(PR [#76](https://github.com/songloft-org/songloft/pull/76) by [@hanxi](https://github.com/hanxi))*
- [`d3cef86`](https://github.com/songloft-org/songloft/commit/d3cef86070306f00fec76e11769e9dc26422ca68) - **agents**: 约定直接提交 main 分支,不建功能分支 *(commit by [@hanxi](https://github.com/hanxi))*
- [`1526608`](https://github.com/songloft-org/songloft/commit/15266082adfdaf5b872504bd49f077c372ebf4ba) - **js-plugin**: 更新客户端 SDK host bridge 适用范围 *(commit by [@hanxi](https://github.com/hanxi))*
- [`87dbec4`](https://github.com/songloft-org/songloft/commit/87dbec40ffdaccca4203ddcb5863ff46e5e2374a) - **repowiki**: 同步 Web 插件页改为 iframe 内嵌 *(commit by [@hanxi](https://github.com/hanxi))*
- [`4a7fbec`](https://github.com/songloft-org/songloft/commit/4a7fbeca50f2ffa50697cda8c94641277da89426) - 全面同步文档与代码并将 repowiki 改为手动维护 *(commit by [@hanxi](https://github.com/hanxi))*
- [`18135f7`](https://github.com/songloft-org/songloft/commit/18135f7451e2c5bf5ab26df082bb8fc32b4163e1) - **repowiki**: 接入文档站并补全英文版 *(commit by [@hanxi](https://github.com/hanxi))*
- [`102936c`](https://github.com/songloft-org/songloft/commit/102936c5b84d644397da0c83e54662bb58108aa9) - **repowiki**: 将 file:// 源码链接改为 GitHub blob 链接 *(commit by [@hanxi](https://github.com/hanxi))*
- [`bbe3ab0`](https://github.com/songloft-org/songloft/commit/bbe3ab0d36c5004bc887e5b254f6594d16684a97) - 同步文档为原生平台统一 media_kit 后端 + 更新子模块指针 (songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`86a70a4`](https://github.com/songloft-org/songloft/commit/86a70a4eed5948eb8777237902f2875a81609b66) - 声明项目纯为爱发电、无收费/赞助渠道，提醒防诈骗 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`c724aa7`](https://github.com/songloft-org/songloft/commit/c724aa746ed3b5f535d94bff008deb5b05a944a0) - **addon**: sync HA add-on version to 2.10.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`6e3e285`](https://github.com/songloft-org/songloft/commit/6e3e285309fb3ad6571d25972620e5696b9de210) - 同步 songloft-player 子模块 (web 端 hls.js 电台修复 songloft-org/songloft[#275](https://github.com/songloft-org/songloft/pull/275)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`c19b54b`](https://github.com/songloft-org/songloft/commit/c19b54b1b5cfcf6794d84957b6b2c56f65668e26) - 同步 songloft-player 子模块 (Windows 退出/播放/日志修复 songloft-org/songloft[#271](https://github.com/songloft-org/songloft/pull/271)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7dd8a1c`](https://github.com/songloft-org/songloft/commit/7dd8a1ccc70e06de1a342d6fc47160dec24dd048) - 更新 songloft-player 子模块指针（分类页播放全部+多选） *(commit by [@hanxi](https://github.com/hanxi))*
- [`76442a6`](https://github.com/songloft-org/songloft/commit/76442a6b395266b9738824d79634f96260dbfd9d) - 更新 miot 子模块指针（补搜索源注册开发者文档） *(commit by [@hanxi](https://github.com/hanxi))*
- [`869a06d`](https://github.com/songloft-org/songloft/commit/869a06d29cc3c370c7742ce16ed55ab0e871c683) - 更新 songloft-player 子模块指针（桌面播放快捷键 [#279](https://github.com/songloft-org/songloft/pull/279)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`edf3e72`](https://github.com/songloft-org/songloft/commit/edf3e720ab49fdf3a23254edf8f8d08249759dfd) - 更新 songloft-player 子模块指针（全屏播放页快捷键修复 [#279](https://github.com/songloft-org/songloft/pull/279)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`fcf741b`](https://github.com/songloft-org/songloft/commit/fcf741b9af967c897a2f77b91129b194f62d85b6) - 更新 songloft-player 子模块指针（打开后自动进入全屏歌词 songloft-org/songloft-player[#19](https://github.com/songloft-org/songloft/pull/19)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`4b453ce`](https://github.com/songloft-org/songloft/commit/4b453ce10d8c87a9846341fb455b112f4d0bbe99) - 更新 songloft-player 子模块指针(插件 Tab 关闭 Android Hybrid Composition 修复菜单栏黑屏 [#273](https://github.com/songloft-org/songloft/pull/273)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`2c5bded`](https://github.com/songloft-org/songloft/commit/2c5bded39aa606d0f4920c926fb2530c9ef10413) - 更新 songloft-player 子模块指针(libmpv 日志接入 FileLogger 排查桌面端 HLS 电台失败 [#249](https://github.com/songloft-org/songloft/pull/249)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`0026796`](https://github.com/songloft-org/songloft/commit/0026796a34d7645da07e066cd516ba0d274d65b5) - 更新 songloft-player 子模块指针(桌面端应用内视频画面渲染 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`64a51ab`](https://github.com/songloft-org/songloft/commit/64a51ab4ea7f18e0e0d66351644b28c2d1494d08) - 更新 songloft-player 子模块指针(侧边/TV 播放器接入视频画面 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3945bf1`](https://github.com/songloft-org/songloft/commit/3945bf1d4508bfd577d7f1132943e941d273cdc2) - 更新 songloft-player 子模块指针(macOS/移动/Web 视频画面 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`ab42a92`](https://github.com/songloft-org/songloft/commit/ab42a92441a1afc1b67f35dd7fc69e0eedb1aca0) - 更新 songloft-player 子模块指针(macOS/移动默认启用 media_kit) + FAQ 视频画面全平台默认支持 (songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`eada187`](https://github.com/songloft-org/songloft/commit/eada1875d743ae82ef3016a2e143437f793a20c6) - 更新 songloft-player 子模块指针(视频画面架构文档 + Web 后端决策记录 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`9c0c1b2`](https://github.com/songloft-org/songloft/commit/9c0c1b29d8f128c8a7957af58edb7ea8b3434a60) - 更新 songloft-player 子模块指针(移除测试未使用 import) *(commit by [@hanxi](https://github.com/hanxi))*
- [`051663b`](https://github.com/songloft-org/songloft/commit/051663b72a441702fc63b88ba9640e127e1b5409) - 更新 songloft-player 子模块指针(添加歌曲/电台是否视频开关 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`2b4d17f`](https://github.com/songloft-org/songloft/commit/2b4d17fb86987693a702ac2cf6e001ba0be0fe6e) - 更新 songloft-player 子模块指针(修复 macOS 视频黑屏与首次播放失败 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`8bc9950`](https://github.com/songloft-org/songloft/commit/8bc995049d5fd4b5446c7983119bea1a2f0a838a) - 更新 songloft-player 子模块指针(单曲循环修复 songloft-org/songloft[#284](https://github.com/songloft-org/songloft/pull/284)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`087ca92`](https://github.com/songloft-org/songloft/commit/087ca92959495539e581670f7ee2c128e6aedbbd) - 更新 songloft-player 子模块指针(TerminateProcess 根治退出报警框 songloft-org/songloft[#271](https://github.com/songloft-org/songloft/pull/271)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`fa1217b`](https://github.com/songloft-org/songloft/commit/fa1217bb9082b6cadfb2e151b22900519a0420b8) - 更新 songloft-player 子模块指针(自定义视图入口收进更多菜单) *(commit by [@hanxi](https://github.com/hanxi))*
- [`cfd8f01`](https://github.com/songloft-org/songloft/commit/cfd8f01a690b2d9a597e5194ab436d40effa8260) - 更新 plugin-toolchain 子模块指针(SDK v2.12.1 + 修复发布 workflow) *(commit by [@hanxi](https://github.com/hanxi))*
- [`df63057`](https://github.com/songloft-org/songloft/commit/df6305754021b4233d7e992fa1b80c642fe661e7) - 更新 songloft-player 子模块指针(歌曲行统一为共享 SongTile) *(commit by [@hanxi](https://github.com/hanxi))*
- [`173bd83`](https://github.com/songloft-org/songloft/commit/173bd834810112c2804d7f760258db662c71e349) - 更新 songloft-player 子模块指针(facet 卡片加播放按钮) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7808602`](https://github.com/songloft-org/songloft/commit/78086027f445357bf76a4b0aa4a6c0d6ca085155) - 更新 plugin-toolchain 子模块指针(v2.12.2:修复 Vue 模板 dev 死循环与重复安装) *(commit by [@hanxi](https://github.com/hanxi))*
- [`f2ca805`](https://github.com/songloft-org/songloft/commit/f2ca805429f47c3fbec23951298e95a9e5899a3e) - 更新 songloft-player 子模块指针(歌单视图工具栏上移顶部 AppBar) *(commit by [@hanxi](https://github.com/hanxi))*
- [`143cc44`](https://github.com/songloft-org/songloft/commit/143cc44984ace371b86d4f7642d218a8e0566220) - 更新 songloft-player 子模块指针(歌单搜索提示词修正) *(commit by [@hanxi](https://github.com/hanxi))*
- [`bdf69b3`](https://github.com/songloft-org/songloft/commit/bdf69b3f347b4be232d17fe5048272d7e789c2ce) - 更新 songloft-player 子模块指针(修复切换歌单子视图 GlobalKey 报错) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7bb2946`](https://github.com/songloft-org/songloft/commit/7bb29466f5bffc7fab4ce7f6bf2b4f38f310d136) - 更新 songloft-player 子模块指针(WebView2 环境/托盘残留/电台直连) *(commit by [@hanxi](https://github.com/hanxi))*
- [`9854627`](https://github.com/songloft-org/songloft/commit/9854627e7c322bddea2b70cc478b6d7f1e489f7f) - 更新 songloft-player 子模块指针(切后台黑屏无 reload 修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`23eba04`](https://github.com/songloft-org/songloft/commit/23eba0444e90705d92493604ff97d39aee5b4cea) - 更新 songloft-player 子模块指针(切后台白屏 Dart 侧重绘修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`dc6a229`](https://github.com/songloft-org/songloft/commit/dc6a229e2b115b3a265e820e4ec28d6ae0e48259) - 更新 songloft-player 子模块指针(切后台白屏补发 webglcontextlost 修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`f3ba3fe`](https://github.com/songloft-org/songloft/commit/f3ba3fe621a53fe3327c442f482d8188df90b547) - 更新 songloft-player 子模块指针(悬浮 console 调试面板) *(commit by [@hanxi](https://github.com/hanxi))*
- [`cf25c55`](https://github.com/songloft-org/songloft/commit/cf25c5506ddf5abeec543e17be44d1b2d5c282ac) - 更新 songloft-player 子模块指针(悬浮 console 增强错误抓取) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e603860`](https://github.com/songloft-org/songloft/commit/e6038606ffc864a54c09ced23c38eaa62f8e71c5) - 更新 songloft-player 子模块指针(TV D-pad 焦点专属布局 + isTv 门控修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3a48de9`](https://github.com/songloft-org/songloft/commit/3a48de97202167774d6cf7553592a7a41c96ccae) - 更新 songloft-player 子模块指针(收尾清理 MultiSurface/恢复 band-aid) *(commit by [@hanxi](https://github.com/hanxi))*
- [`42d899c`](https://github.com/songloft-org/songloft/commit/42d899cf2811ac16ca937b2d4042263581a3d81f) - 更新 songloft-player 子模块指针(onReorder→onReorderItem 迁移,Flutter 3.44.6) *(commit by [@hanxi](https://github.com/hanxi))*
- [`79e2fd9`](https://github.com/songloft-org/songloft/commit/79e2fd9b3e551ef2fe7b4947e7d2d5a619eca617) - 更新 songloft-player 子模块指针(修复最小化残留 HWND 拦截桌面右键 [#293](https://github.com/songloft-org/songloft/pull/293)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`9e99919`](https://github.com/songloft-org/songloft/commit/9e99919febb521326b5c13978fabc05da791f9d8) - 更新 songloft-player 子模块至逐字歌词版本 *(PR [#294](https://github.com/songloft-org/songloft/pull/294) by [@hanxi](https://github.com/hanxi))*
- [`a2b98b9`](https://github.com/songloft-org/songloft/commit/a2b98b98f9389e68f70bbf8d3a944f0f26410702) - update miot plugin submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`a833d63`](https://github.com/songloft-org/songloft/commit/a833d63625f1ca1533986016260f4483ddc8ab6c) - update miot plugin submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`a5e5d08`](https://github.com/songloft-org/songloft/commit/a5e5d080ae6230bc656a4505d480e4c22ecf50bd) - 更新 songloft-player 子模块 (修复 Web 语义节点遮挡插件 iframe [#295](https://github.com/songloft-org/songloft/pull/295)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7fc4cbe`](https://github.com/songloft-org/songloft/commit/7fc4cbe5cd19acc4324990bcfccf751ec393b0ef) - 更新 songloft-player 子模块 (修复移动端 media_kit 后端 Android 全曲无法播放 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`2ae3be6`](https://github.com/songloft-org/songloft/commit/2ae3be6b58b235c86ef447a30302c03873f3da42) - 更新 songloft-player 子模块 (media_kit 后端实现 setAndroidAudioAttributes，修复 Android 全曲无法播放 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`f2f9afa`](https://github.com/songloft-org/songloft/commit/f2f9afaf38ea10cad7dc05a8e19094d3595c8c76) - 更新 songloft-player 子模块 (原生平台统一 media_kit 后端，移除原生后端回退 songloft-org/songloft[#76](https://github.com/songloft-org/songloft/pull/76)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b46fe0f`](https://github.com/songloft-org/songloft/commit/b46fe0f80aad05f1b172a6c05db2b25497499820) - 更新 songloft-player 子模块（修复 NextConsole UMD 全局实例化，Web 调试面板悬浮按钮恢复显示） *(commit by [@hanxi](https://github.com/hanxi))*
- [`b411685`](https://github.com/songloft-org/songloft/commit/b41168506d2b3c143f2ce7038aa9f217aaee5ead) - 更新 songloft-player 子模块（修复 Web 切后台回来封面偶发纯黑，resume 驱逐图片缓存+重解码） *(commit by [@hanxi](https://github.com/hanxi))*
- [`d08205b`](https://github.com/songloft-org/songloft/commit/d08205b55b6b58febabd5a5bec2c8df44250b901) - 更新 songloft-player 子模块（修复 Web 应用内切页面回来列表封面纯黑，导航时驱逐图片缓存重解码） *(commit by [@hanxi](https://github.com/hanxi))*
- [`3137005`](https://github.com/songloft-org/songloft/commit/3137005065e376c317ed213135eae808ebc26dde) - 更新 songloft-player 子模块（修复 Web 封面切插件 WebView 页返回后变黑，按封面精准驱逐缓存重解码） *(commit by [@hanxi](https://github.com/hanxi))*
- [`eb4c9ba`](https://github.com/songloft-org/songloft/commit/eb4c9ba8ae5f6e70a4d73a2601243062e1361601) - 更新 songloft-player 子模块（Web 强制 CanvasKit CPU 渲染根治封面偶发纯黑，回退应用层缓解） *(commit by [@hanxi](https://github.com/hanxi))*
- [`adecd8b`](https://github.com/songloft-org/songloft/commit/adecd8b4851098a5264a6dad2cb68a5c7c4a48bd) - 更新 songloft-player 子模块（封面切tab/筛选丢失根治：调大imageCache+缩略解码，回退CPU渲染） *(commit by [@hanxi](https://github.com/hanxi))*
- [`80bd72f`](https://github.com/songloft-org/songloft/commit/80bd72fadb5486121cc56c65c681c8be3212d170) - 更新 songloft-player 子模块（CoverImage 加临时诊断日志排查封面三态） *(commit by [@hanxi](https://github.com/hanxi))*
- [`3f74c1e`](https://github.com/songloft-org/songloft/commit/3f74c1e474d25995a35684ca021b181642743d39) - 更新 songloft-player 子模块（Web 封面走 HttpGet 解码缩小 GPU 纹理 + 删冗余 imageBuilder + 加 WebGL context 丢失诊断日志） *(commit by [@hanxi](https://github.com/hanxi))*
- [`df4e1d4`](https://github.com/songloft-org/songloft/commit/df4e1d4d757e97d2417b1231b7565b6c59c9f118) - 更新 songloft-player 子模块（所有封面卡片统一走 HttpGet 缩略解码，修专辑/歌手/歌单列表封面黑） *(commit by [@hanxi](https://github.com/hanxi))*
- [`1df7170`](https://github.com/songloft-org/songloft/commit/1df7170778c753047ebc56d79a423a4ae21736e6) - 更新 songloft-player 子模块（清理封面诊断日志 + 文档记录 CanvasKit 大纹理封面变黑坑） *(commit by [@hanxi](https://github.com/hanxi))*
- [`05601ea`](https://github.com/songloft-org/songloft/commit/05601ead34d58bbc85b25457ae38d10b0566ee5d) - 更新 songloft-player 子模块（修复歌单单曲播放队列被截断到已加载页 [#299](https://github.com/songloft-org/songloft/pull/299)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`0094f1a`](https://github.com/songloft-org/songloft/commit/0094f1a64f20e795e801ad8ef15e34aba095b7d9) - 更新 songloft-player 子模块（修复分类页点击单曲播放队列被截断 [#299](https://github.com/songloft-org/songloft/pull/299)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`f6c87d6`](https://github.com/songloft-org/songloft/commit/f6c87d6ea3af260838e92056de53d41dff1824d6) - 更新 songloft-player 子模块（文档记录分页列表点单曲播放队列构建铁律 [#299](https://github.com/songloft-org/songloft/pull/299)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`55ca279`](https://github.com/songloft-org/songloft/commit/55ca2791e2aa09dbab1a8fb93e75bc59153f5f23) - 更新 downloader 子模块（按歌单/艺术家/专辑筛选下载 [#304](https://github.com/songloft-org/songloft/pull/304)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`d124f88`](https://github.com/songloft-org/songloft/commit/d124f88cfed9ee2d9696fdad24530ebb40be5aac) - 更新 songloft-player 子模块（桌面折叠播放栏补齐拖动滑块 [#301](https://github.com/songloft-org/songloft/pull/301)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`1ba55a4`](https://github.com/songloft-org/songloft/commit/1ba55a4260cde699b2df0b89d80d21bd704a0316) - 更新 songloft-player 子模块（插件 iframe 抖动诊断埋点 [#278](https://github.com/songloft-org/songloft/pull/278)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`1233bf5`](https://github.com/songloft-org/songloft/commit/1233bf531e315b0efcac96cdc24f2ee2d6f1bd7b) - 更新 songloft-player 子模块（设置页显示后端构建时间） *(commit by [@hanxi](https://github.com/hanxi))*
- [`cb64286`](https://github.com/songloft-org/songloft/commit/cb64286b20af55a9d27c559d20afe1502dd412f0) - 更新 songloft-player 子模块（加整页抖动尺寸振荡探针 [#278](https://github.com/songloft-org/songloft/pull/278)） *(commit by [@hanxi](https://github.com/hanxi))*
- [`d60d15a`](https://github.com/songloft-org/songloft/commit/d60d15a2ebcd4ec3aff97066347ac12096c66143) - release version 2.11.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.10.0] - 2026-07-11
### :sparkles: New Features
- [`510f309`](https://github.com/songloft-org/songloft/commit/510f3098d7fd5f41883882c2c287b8bc08b17033) - **songs**: 支持按歌单 label 排除歌曲，默认排除隐藏歌单 songloft-org/songloft-player[#18](https://github.com/songloft-org/songloft/pull/18) *(commit by [@hanxi](https://github.com/hanxi))*
- [`cf84ae0`](https://github.com/songloft-org/songloft/commit/cf84ae05bfe0c7ab76dbdafa4d11c9f7704097d0) - **addon**: 新增 Home Assistant 加载项 *(commit by [@hanxi](https://github.com/hanxi))*
- [`7e938d0`](https://github.com/songloft-org/songloft/commit/7e938d0cf4a31758683e928b3471d56ba0d1bc88) - **organize**: 目录整理新增 preview 与插件 bridge，修复 file_path 双前缀 *(PR [#261](https://github.com/songloft-org/songloft/pull/261) by [@hanxi](https://github.com/hanxi))*
- [`b6b42c4`](https://github.com/songloft-org/songloft/commit/b6b42c4e5fa41d92d17dda14501e687021660df7) - **metadata**: 网络歌曲导入即探测时长，修复音箱无法自动切歌 *(PR [#264](https://github.com/songloft-org/songloft/pull/264) by [@hanxi](https://github.com/hanxi))*
- [`90c964e`](https://github.com/songloft-org/songloft/commit/90c964e18bcec0a94ef962048aaebc451a236279) - **scan**: 目录级定向扫描，过期清理按作用域收敛 *(PR [#262](https://github.com/songloft-org/songloft/pull/262) by [@hanxi](https://github.com/hanxi))*
- [`617ff39`](https://github.com/songloft-org/songloft/commit/617ff3928866bbfcf31124f794db3713c2979702) - **songs**: 支持编辑本地歌曲改名，修复插件音源歌曲编辑 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e377144`](https://github.com/songloft-org/songloft/commit/e3771443b7d5367742afd8a11e0d663b2f2ea577) - **tags**: WriteSongTags 支持 track 音轨号字段 *(PR [#269](https://github.com/songloft-org/songloft/pull/269) by [@hanxi](https://github.com/hanxi))*
- [`0a00903`](https://github.com/songloft-org/songloft/commit/0a00903e0d833a810548b27ee44d51ff6a1c9e2a) - **scan**: 支持 mov 格式音频 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`26aa89f`](https://github.com/songloft-org/songloft/commit/26aa89f0f76f80113cd94b341596b7d601051c9d) - **miot**: update plugin for touchscreen lyrics playback *(commit by [@hanxi](https://github.com/hanxi))*
- [`1f4f57e`](https://github.com/songloft-org/songloft/commit/1f4f57ef76a59a8ee4b12e5ecc32d4a05116cd62) - **radio**: prevent stuck radio streams *(commit by [@hanxi](https://github.com/hanxi))*
- [`e1efc32`](https://github.com/songloft-org/songloft/commit/e1efc32464d2ebd74f38deb906461e48fb621cb8) - **upgrade**: enforce matching update channels *(commit by [@hanxi](https://github.com/hanxi))*
- [`99152f0`](https://github.com/songloft-org/songloft/commit/99152f0e7ba74aee2545b874752c5710f09d2fea) - **jsplugin**: deliver host events during awaits *(commit by [@hanxi](https://github.com/hanxi))*
- [`87be549`](https://github.com/songloft-org/songloft/commit/87be549507d67c68afda018f6b3e02a544cf13ff) - **release**: repair bundled macos signing *(commit by [@hanxi](https://github.com/hanxi))*
- [`11fb982`](https://github.com/songloft-org/songloft/commit/11fb982f90919db8df47654195a97fc8f2093180) - **docs**: 修复英文 README 同步后 docs/en/ 链接多套一层导致死链 *(commit by [@hanxi](https://github.com/hanxi))*
- [`adeb8b0`](https://github.com/songloft-org/songloft/commit/adeb8b0166083bbce5a90b28f42cf7a602fb1a2d) - **jsplugin**: sync plugin toolchain permissions *(commit by [@hanxi](https://github.com/hanxi))*
- [`2e0b21f`](https://github.com/songloft-org/songloft/commit/2e0b21f1652aa4e5411c4b7cc156f70a6ba540eb) - **cover**: prevent slow cover requests from hanging *(commit by [@hanxi](https://github.com/hanxi))*
- [`c60bebb`](https://github.com/songloft-org/songloft/commit/c60bebb2ac99e8485fb9693b3f64083da20e6da2) - **cover**: 修复自动歌单封面重新扫描后随机变化 *(commit by [@hanxi](https://github.com/hanxi))*
- [`0774040`](https://github.com/songloft-org/songloft/commit/07740402257e7c4b23b02d76cf461d06baffba04) - **server**: 支持 port=0 由系统自动分配监听端口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`09c569e`](https://github.com/songloft-org/songloft/commit/09c569ed0681aaf03ca303f5cfa7ed573691d865) - **ci**: mac bundle 兜底重签给 songloft-server 补 inherit entitlements *(commit by [@hanxi](https://github.com/hanxi))*
- [`2c1ae52`](https://github.com/songloft-org/songloft/commit/2c1ae5298c6cb0efe9d27bc70f03e38c02f81dd0) - **server**: 新增 -music flag 供 Bundle 桌面模式传入音乐目录 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a4b209d`](https://github.com/songloft-org/songloft/commit/a4b209dc696f5cd89fd76972188ae87b43d2fff2) - **playlist**: 隐藏歌单后支持可见子集重排，修复排序报错 *(PR [#266](https://github.com/songloft-org/songloft/pull/266) by [@hanxi](https://github.com/hanxi))*
- [`66b2e62`](https://github.com/songloft-org/songloft/commit/66b2e62fcaef8638280e43effc8ee5dfcc3b1c88) - **upgrade**: 检查更新加短超时并持久化 GitHub 代理 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ffabd30`](https://github.com/songloft-org/songloft/commit/ffabd30889bcb088573df073e4d3c7aa512ed473) - **release**: package valid bundled iOS app *(commit by [@hanxi](https://github.com/hanxi))*

### :zap: Performance Improvements
- [`43fdfee`](https://github.com/songloft-org/songloft/commit/43fdfee0667dccc6d89e0b9accfd81ea48cd09d0) - **scan**: 修复 CUE 整轨切分对大 CD 镜像退化为 O(N²) 慢速 seek *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`756d048`](https://github.com/songloft-org/songloft/commit/756d04876546ff4bf867eca1bfcd6192eb6b4ded) - update CHANGELOG for v2.9.6 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`f8858fb`](https://github.com/songloft-org/songloft/commit/f8858fba04daa40f96d05cda147073311231903a) - 重构文档网站首页 *(commit by [@hanxi](https://github.com/hanxi))*
- [`9a7d875`](https://github.com/songloft-org/songloft/commit/9a7d875383f189a69887d4e4b17da8d4b73a1a54) - 添加截图
- [`08bc3d2`](https://github.com/songloft-org/songloft/commit/08bc3d2d61f1fff690660a817159a7473255c4c8) - 新增界面截图一览并在 README 补充截图与语言切换 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e63fcf8`](https://github.com/songloft-org/songloft/commit/e63fcf8c9ea3ed6c5b603ef14693d9aac5cc9c0b) - **faq**: 说明 Firefox 下 Web 端点击偏移的兼容性限制 *(commit by [@hanxi](https://github.com/hanxi))*
- [`555fb12`](https://github.com/songloft-org/songloft/commit/555fb1253f33e51e95050cf15694045b1aeb5543) - **agents**: 补充 addon 目录说明与文档站结构章节 *(commit by [@hanxi](https://github.com/hanxi))*
- [`d25d8a6`](https://github.com/songloft-org/songloft/commit/d25d8a66524a001e496809cc331ab5588d0bf551) - Windows 客户端支持 Scoop 安装 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`f866fb4`](https://github.com/songloft-org/songloft/commit/f866fb450e28e72bc5a9fa8193696f25d6f67956) - **miot**: update plugin submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`1b66d48`](https://github.com/songloft-org/songloft/commit/1b66d484acf4c3fe279a4acd34dfb3f2e9907a82) - **plugins**: update miot submodule for search priority (songloft-org/songloft-plugin-miot[#30](https://github.com/songloft-org/songloft/pull/30)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`d69ef52`](https://github.com/songloft-org/songloft/commit/d69ef529017f75da5862eb623310a8c2bea7742f) - **player**: update auth wallet fix *(commit by [@hanxi](https://github.com/hanxi))*
- [`30c1d02`](https://github.com/songloft-org/songloft/commit/30c1d021db68e633dd93ca517729d942ebc26b5a) - **submodules**: update player and radio plugin refs *(commit by [@hanxi](https://github.com/hanxi))*
- [`a8c730f`](https://github.com/songloft-org/songloft/commit/a8c730f222faffcf9f0f608b56a2b1de08d08f6d) - **player**: update submodule for Windows HLS fix [#249](https://github.com/songloft-org/songloft/pull/249) *(commit by [@hanxi](https://github.com/hanxi))*
- [`839606a`](https://github.com/songloft-org/songloft/commit/839606a94fc6937a761547f9b3ec871268cfb822) - **player**: update submodule for hidden-playlist song filtering songloft-org/songloft-player[#18](https://github.com/songloft-org/songloft/pull/18) *(commit by [@hanxi](https://github.com/hanxi))*
- [`6dba52d`](https://github.com/songloft-org/songloft/commit/6dba52df56ffb86b7a699b9988a44bb919e5671d) - **deps**: 更新 songloft-player 修复 mac 本地模式 *(commit by [@hanxi](https://github.com/hanxi))*
- [`b173f91`](https://github.com/songloft-org/songloft/commit/b173f91bf557f91a756f09a2e731dd7428800fec) - **deps**: 更新 songloft-player 修复 iOS 本地模式 *(commit by [@hanxi](https://github.com/hanxi))*
- [`eb934af`](https://github.com/songloft-org/songloft/commit/eb934af0ce834a88b8576d10f31c0c281c97a467) - **deps**: 更新 songloft-player 修复退出登录与桌面本地扫描 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f3886cf`](https://github.com/songloft-org/songloft/commit/f3886cfcffea12079375fe630667ef3747887733) - **deps**: 更新 songloft-player 支持编辑本地歌曲与修复编辑入口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`11e9ff4`](https://github.com/songloft-org/songloft/commit/11e9ff4f173d283dc71275b1be42b1b1f5ac178b) - **player**: 更新子模块，记住 GitHub 代理并抽取选择 mixin *(commit by [@hanxi](https://github.com/hanxi))*
- [`0890f59`](https://github.com/songloft-org/songloft/commit/0890f5933e564d21e3b43749e4e617d5ece01996) - **player**: 更新子模块，统一 GitHub 代理选择为顶部下拉 *(commit by [@hanxi](https://github.com/hanxi))*
- [`607e381`](https://github.com/songloft-org/songloft/commit/607e381f44706ff7ada9a6781d5b3ccd0e61799a) - **deps**: 更新 songloft-player 修复 iOS 本地模式扫描沙盒报错 *(commit by [@hanxi](https://github.com/hanxi))*
- [`b607701`](https://github.com/songloft-org/songloft/commit/b607701465cc11f41e033363cdd483063ed8aad8) - release version 2.10.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.6] - 2026-07-04
### :sparkles: New Features
- [`8657599`](https://github.com/songloft-org/songloft/commit/865759926a5c7835a7ad9787ae3ee17b34422851) - **jsruntime**: add AES decrypt bridge *(PR [#248](https://github.com/songloft-org/songloft/pull/248) by [@fly818](https://github.com/fly818))*

### :bug: Bug Fixes
- [`846a19a`](https://github.com/songloft-org/songloft/commit/846a19a0cbad9731d19d3fb961cbb9ea9ae72087) - **miot**: update auto-next playback fix *(commit by [@hanxi](https://github.com/hanxi))*
- [`d2305ff`](https://github.com/songloft-org/songloft/commit/d2305ffa3a9ebb32687e54550b7cb611369a5963) - **player**: 更新 Windows 客户端退出修复 [#246](https://github.com/songloft-org/songloft/pull/246) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`e2bcdcb`](https://github.com/songloft-org/songloft/commit/e2bcdcb4be69954011654acd4cff067889e13dfb) - update CHANGELOG for v2.9.5 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`349dfdb`](https://github.com/songloft-org/songloft/commit/349dfdb8654f762da7aa493a41d8e25755bf1aa5) - **jsruntime**: document AES decrypt follow-ups *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`0bd87dd`](https://github.com/songloft-org/songloft/commit/0bd87ddfbb0ab86ff3692e3e6e8fcefe4155f1da) - **submodules**: update miot plugin and toolchain *(commit by [@hanxi](https://github.com/hanxi))*
- [`07e089b`](https://github.com/songloft-org/songloft/commit/07e089bbead0f1961ecf90dcbccc005ff5b85d78) - **submodules**: update miot plugin *(commit by [@hanxi](https://github.com/hanxi))*
- [`2810ceb`](https://github.com/songloft-org/songloft/commit/2810ceb9ff7603ca57e72201fd0b8c18cc6b0b76) - **submodules**: update plugin-toolchain *(commit by [@hanxi](https://github.com/hanxi))*
- [`c3a3be6`](https://github.com/songloft-org/songloft/commit/c3a3be69de908d15ddb53bbe434e7fcc11a55da8) - release version 2.9.6 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.5] - 2026-07-03
### :sparkles: New Features
- [`836ddc7`](https://github.com/songloft-org/songloft/commit/836ddc76297e7e6b8641086b130090c8edd59c30) - CUE Sheet 整轨音乐支持 *(PR [#33](https://github.com/songloft-org/songloft/pull/33) by [@hanxi](https://github.com/hanxi))*
- [`850c996`](https://github.com/songloft-org/songloft/commit/850c9963f7b03a1a89c029b1ebc6f92e9f516804) - 歌曲支持按文件修改时间排序 [#242](https://github.com/songloft-org/songloft/pull/242) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b29d052`](https://github.com/songloft-org/songloft/commit/b29d0521d6a446a35a5de24279fa428693631a1f) - **jsplugin**: support inbound websocket handlers *(commit by [@hanxi](https://github.com/hanxi))*
- [`e24c2e2`](https://github.com/songloft-org/songloft/commit/e24c2e2678c5e40b869950d8801cce95aa40ea13) - **jsruntime**: 新增原生 __go_crypto_sha256_bytes 与 __go_crypto_rc4 *(commit by [@hanxi](https://github.com/hanxi))*
- [`da9158f`](https://github.com/songloft-org/songloft/commit/da9158ff78edc10eba32bb4d26939e546591e24d) - **jsruntime**: 新增原生 __go_crypto_sha1 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e2e6206`](https://github.com/songloft-org/songloft/commit/e2e620644e3bb8b2b5068c416970734154f280cd) - **source**: 插件音源可返回自定义请求头并在代理/下载时应用 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`46c57dc`](https://github.com/songloft-org/songloft/commit/46c57dc7375c698c0cd2c4f1ff392dedce575895) - 修复内存泄漏和性能问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`3ce474d`](https://github.com/songloft-org/songloft/commit/3ce474d77f1a278a4d8b91f98844c224c24ba64f) - auto-create 歌单按名复用 ID 避免每次扫描重建 *(commit by [@hanxi](https://github.com/hanxi))*
- [`0e2bfef`](https://github.com/songloft-org/songloft/commit/0e2bfef9efbc61446fcfeca8d24999c8ff15249b) - **jsruntime**: support binary fetch payloads *(commit by [@hanxi](https://github.com/hanxi))*

### :zap: Performance Improvements
- [`4a55fc2`](https://github.com/songloft-org/songloft/commit/4a55fc2ad8cca8957446a9249c38d3b6e7e463c8) - **jsruntime**: 优化 JS 插件热路径与冷启动 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`e9d03e5`](https://github.com/songloft-org/songloft/commit/e9d03e57513903b96b8306ef2bf9184a286dbf5a) - update CHANGELOG for v2.9.4 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`95575d9`](https://github.com/songloft-org/songloft/commit/95575d92c24841d1371fad50cf662d8ce09e9d21) - update pkg/tag submodule (FLAC CUESHEET 支持) *(commit by [@hanxi](https://github.com/hanxi))*
- [`28bdb36`](https://github.com/songloft-org/songloft/commit/28bdb36b3a9c873952a754f4da1937879f0cf620) - update songloft-plugin-miot submodule (关闭口令修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`56d08d6`](https://github.com/songloft-org/songloft/commit/56d08d61dd19989eb07c2110bbd22665d16cbc74) - update songloft-plugin-miot submodule (搜索接口选项修复) *(commit by [@hanxi](https://github.com/hanxi))*
- [`6dd6cc6`](https://github.com/songloft-org/songloft/commit/6dd6cc6b20abe5635e1f8f82c11f72936a3a3006) - update songloft-plugin-miot submodule (歌单 ID 失效重试) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7994aed`](https://github.com/songloft-org/songloft/commit/7994aed16662c01cdca10e7d0e444481c00d47d9) - update songloft-plugin-subsonic submodule (修复外部歌曲歌词 404) *(commit by [@hanxi](https://github.com/hanxi))*
- [`28614d1`](https://github.com/songloft-org/songloft/commit/28614d15377ba6a19a4a74905bbb46cb4d904c28) - update songloft-plugin-subsonic submodule (bump v2.2.1) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b5201ce`](https://github.com/songloft-org/songloft/commit/b5201cefd8baf67f028fda8d408f4ccd304a83ba) - update songloft-plugin-miot submodule (新增触屏歌词开关 [#239](https://github.com/songloft-org/songloft/pull/239)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7b1e58a`](https://github.com/songloft-org/songloft/commit/7b1e58a717671ff86a32166ea9c6e76b05ab236c) - release version 2.9.5 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.4] - 2026-06-30
### :sparkles: New Features
- [`7da4a96`](https://github.com/songloft-org/songloft/commit/7da4a9686545c7543ae4587ec02b664f842f47dc) - 删除歌曲支持删除本地文件、歌单隐藏功能 [#235](https://github.com/songloft-org/songloft/pull/235) *(commit by [@hanxi](https://github.com/hanxi))*
- [`f28d4ee`](https://github.com/songloft-org/songloft/commit/f28d4eee087b1734524480d94d1bb3fe3fb6e7f9) - **jsplugin**: 新增 UDP socket Bridge API (songloft.net) [#222](https://github.com/songloft-org/songloft/pull/222) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3919550`](https://github.com/songloft-org/songloft/commit/3919550151dd240b9e1d670c750b5a4e68723db2) - **jsplugin**: 插件常驻运行白名单 [#237](https://github.com/songloft-org/songloft/pull/237) *(commit by [@hanxi](https://github.com/hanxi))*

### :white_check_mark: Tests
- [`e0f3b5f`](https://github.com/songloft-org/songloft/commit/e0f3b5f6950330ffa526d77804f446bc723adba2) - **jsplugin**: 新增 UDP socket API 单元测试 + 文档更新 [#222](https://github.com/songloft-org/songloft/pull/222) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`0c3d882`](https://github.com/songloft-org/songloft/commit/0c3d88244510e142ad7657cd78f05811578add4f) - update CHANGELOG for v2.9.3 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`d8b3fd6`](https://github.com/songloft-org/songloft/commit/d8b3fd6a81c36443ec11cbd60b1da8137e6210d6) - 补充 Bundle 本地模式文档 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`edb9249`](https://github.com/songloft-org/songloft/commit/edb9249b773b0b8817c3cfe6e2b9b408aba0c3cb) - release version 2.9.4 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.3] - 2026-06-30
### :bug: Bug Fixes
- [`66a3453`](https://github.com/songloft-org/songloft/commit/66a3453b69ddef57049f7c5a356fdd33a90e9f55) - **tag**: update WAV metadata parser *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`c661b00`](https://github.com/songloft-org/songloft/commit/c661b004f655e0d900c7398c5a190af88f059073) - update CHANGELOG for v2.9.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`91e03e7`](https://github.com/songloft-org/songloft/commit/91e03e7d9184a27b747ab75f9ea36e2d7ff3df4f) - clarify issue references in commits *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`cbb1718`](https://github.com/songloft-org/songloft/commit/cbb17183a74a65a00fc7eb97cdd80a8c578f7507) - release version 2.9.3 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.2] - 2026-06-29
### :memo: Documentation Changes
- [`be5455d`](https://github.com/songloft-org/songloft/commit/be5455df89a905bf2dee0bdb8a1e41ed7f3585c6) - update CHANGELOG for v2.9.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`03ea947`](https://github.com/songloft-org/songloft/commit/03ea947a9025bc0d91da14378f14fe761799251f) - **player**: update songloft-player submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`0047882`](https://github.com/songloft-org/songloft/commit/00478824d3e748ccb0cd0a2d96375aa025345a9a) - release version 2.9.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.1] - 2026-06-29
### :bug: Bug Fixes
- [`5b92b91`](https://github.com/songloft-org/songloft/commit/5b92b91e574641146b7101dcc7a827b650ab0f2d) - **player**: update frontend playback crash fix *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`9534a8e`](https://github.com/songloft-org/songloft/commit/9534a8ea1103dfc1cac47ff28f0ebad3671d873d) - update CHANGELOG for v2.9.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`1af5db9`](https://github.com/songloft-org/songloft/commit/1af5db96a6512e388e018f774a54b8a32551852a) - release version 2.9.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.9.0] - 2026-06-29
### :sparkles: New Features
- [`993a3e7`](https://github.com/songloft-org/songloft/commit/993a3e742d9ba5ad7150e3adcd90863b43fec2d0) - **scan**: 歌单创建方式支持三种模式（按文件夹/按顶层文件夹/包含子目录） *(commit by [@hanxi](https://github.com/hanxi))*
- [`e858260`](https://github.com/songloft-org/songloft/commit/e858260ad4f662bb54faef99ea618456a83c88fe) - **plugin**: 新增自动下载 bridge API，支持插件注册缓存完成后自动下载 *(PR [#224](https://github.com/songloft-org/songloft/pull/224) by [@hanxi](https://github.com/hanxi))*
- [`2989703`](https://github.com/songloft-org/songloft/commit/2989703dc32c1daf8280659dd1751c795bee1964) - **equalizer**: 添加 EQ 均衡器 settings 端点 *(PR [#217](https://github.com/songloft-org/songloft/pull/217) by [@hanxi](https://github.com/hanxi))*
- [`8f82ef7`](https://github.com/songloft-org/songloft/commit/8f82ef758af4f58b32f8f9ed78b781cc8bea876d) - 将 Go 后端嵌入客户端，支持本地模式播放 *(PR [#225](https://github.com/songloft-org/songloft/pull/225) by [@hanxi](https://github.com/hanxi))*
- [`7cf670d`](https://github.com/songloft-org/songloft/commit/7cf670de00918bac88be41adb2b0f50fff01d018) - **plugin**: 插件持久化存储 API + 卸载保留数据 + 孤儿清理 *(PR [#220](https://github.com/songloft-org/songloft/pull/220) by [@hanxi](https://github.com/hanxi))*
- [`0f732ed`](https://github.com/songloft-org/songloft/commit/0f732edddf4ac6c979939099bc24c5c0ea2540be) - **player**: 更新子模块实现按服务器隔离凭证 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`0456bbf`](https://github.com/songloft-org/songloft/commit/0456bbf9b6d191903e072c9dde1c0266b260eebe) - **scan**: 修复元数据提取失败时标题含扩展名、空数据覆盖已有记录的问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`02a866b`](https://github.com/songloft-org/songloft/commit/02a866b14c5521b3c4b6696755729724f57b937c) - **cache**: 修复插件来源歌曲缓存/下载时扩展名始终为 .mp3 的问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`0eeb224`](https://github.com/songloft-org/songloft/commit/0eeb2240b274d77f8f5fecfaa5c46073db6335cc) - **cache**: ffprobe format_name 未标准化导致 M4A 文件可能获得 .mov 扩展名 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f3aa113`](https://github.com/songloft-org/songloft/commit/f3aa11359eedc8aac6bc30c8d25634bb0d662364) - **subsonic**: 修复 Subsonic 服务端模式无法播放远程歌曲和电台 *(PR [#219](https://github.com/songloft-org/songloft/pull/219) by [@hanxi](https://github.com/hanxi))*
- [`c5f2cb2`](https://github.com/songloft-org/songloft/commit/c5f2cb22b8840dfd5bdfb4388e5848a8c31eec54) - **scan**: 扫描完成后返回本地歌曲总数，解决与歌曲库计数不一致的困惑 *(commit by [@hanxi](https://github.com/hanxi))*
- [`49fff2e`](https://github.com/songloft-org/songloft/commit/49fff2e343cbbf61a0afc9daa4e9646ce50cdd76) - **plugin**: EnablePlugin 前先卸载残留环境，避免 env already exists 错误 *(commit by [@hanxi](https://github.com/hanxi))*
- [`e14cae5`](https://github.com/songloft-org/songloft/commit/e14cae55713ad06224048683e73599bf6bea045d) - **cover**: 封面 URL 追加时间戳参数穿透客户端缓存 *(PR [#218](https://github.com/songloft-org/songloft/pull/218) by [@hanxi](https://github.com/hanxi))*
- [`5e36bf0`](https://github.com/songloft-org/songloft/commit/5e36bf0ba34d74ea916604476a36e2720a9d2ef4) - **ci**: 添加 gomobile tool 依赖修复 iOS/Android 构建 *(commit by [@hanxi](https://github.com/hanxi))*
- [`73e28ec`](https://github.com/songloft-org/songloft/commit/73e28ec345913cc4daf8ca54d23b1c01635e0af6) - **plugin**: playlists.getSongs 桥接增加诊断日志和重试 *(PR [#21](https://github.com/songloft-org/songloft/pull/21) by [@hanxi](https://github.com/hanxi))*
- [`a3811da`](https://github.com/songloft-org/songloft/commit/a3811daeb18655e93c1a4041155c1f39cbac96d2) - **ci**: 更新 songloft-player 子模块修复 macOS 构建 *(commit by [@hanxi](https://github.com/hanxi))*
- [`412b28c`](https://github.com/songloft-org/songloft/commit/412b28cc5d737768c80006da632536e865131e54) - **plugin**: 私有 GitHub 仓库插件通过 API 端点下载 *(commit by [@hanxi](https://github.com/hanxi))*
- [`90f5a79`](https://github.com/songloft-org/songloft/commit/90f5a79fc86ab8386b5044b60162311d8b671c77) - **player**: 更新 songloft-player 子模块修复本地模式 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f812e1d`](https://github.com/songloft-org/songloft/commit/f812e1d943b6e91cd3508f553ed7bdd02792f2ff) - **player**: 更新子模块修复 Xcode 26.5 编译 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a48b776`](https://github.com/songloft-org/songloft/commit/a48b776f15e15b023825a18219c6d3062ead8389) - **player**: 更新子模块修复 macOS linker 错误 *(commit by [@hanxi](https://github.com/hanxi))*
- [`1c27231`](https://github.com/songloft-org/songloft/commit/1c272317ae3738a9a1fef483e49cdb6a288a88c8) - **player**: 更新子模块修复 iOS 构建 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a0b8c1b`](https://github.com/songloft-org/songloft/commit/a0b8c1bfeea37026b1a67be1177ba38cdc824636) - **player**: 更新子模块修复 iOS EqualizerPlugin 和 SongloftBackendPlugin 编译 *(commit by [@hanxi](https://github.com/hanxi))*
- [`da65db1`](https://github.com/songloft-org/songloft/commit/da65db17f6cc1216d0413d85d41842e661b7aee0) - 修复 Android bundled 模式下 covers 目录路径解析错误 *(commit by [@hanxi](https://github.com/hanxi))*
- [`8739e1a`](https://github.com/songloft-org/songloft/commit/8739e1a5f4e36483c43b301355ce3baebbcde5ea) - **local**: 修复 Android 本地模式扫描 + 更新子模块 *(commit by [@hanxi](https://github.com/hanxi))*
- [`039bdf5`](https://github.com/songloft-org/songloft/commit/039bdf5ca454d1146126d712dd1194fb9b10a71a) - **player**: 更新子模块修复本地模式设置页入口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`56e042c`](https://github.com/songloft-org/songloft/commit/56e042c49beda84780ed1e207f585d8d00220ddd) - **scan**: 扫描文件发现阶段增加进度报告 *(PR [#227](https://github.com/songloft-org/songloft/pull/227) by [@hanxi](https://github.com/hanxi))*
- [`281a8d5`](https://github.com/songloft-org/songloft/commit/281a8d59401d4365e60c6d4c27afe93bd03c5d05) - **settings**: 更新子模块修复歌单创建方式下拉竖排显示 *(PR [#228](https://github.com/songloft-org/songloft/pull/228) by [@hanxi](https://github.com/hanxi))*
- [`e5ac251`](https://github.com/songloft-org/songloft/commit/e5ac2518898a07cd3695ba8b1626ed54756c2c60) - **web**: 更新子模块修复移动端浏览器切后台黑屏 *(PR [#229](https://github.com/songloft-org/songloft/pull/229) by [@hanxi](https://github.com/hanxi))*
- [`7c13a31`](https://github.com/songloft-org/songloft/commit/7c13a31e07855ab1e077530ad362a29e10049f43) - **player**: 更新子模块修复桌面端播完不自动切歌 *(commit by [@hanxi](https://github.com/hanxi))*
- [`bd84888`](https://github.com/songloft-org/songloft/commit/bd84888b7793a8c6f7d4a0b89b127891468f5941) - 标签解析与扫描稳定性修复 (songloft-org/songloft-player[#14](https://github.com/songloft-org/songloft/pull/14)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e433aa3`](https://github.com/songloft-org/songloft/commit/e433aa3d3b636e58b56d67c962b3d96552377e23) - **jsplugin**: serveFile 支持非本地歌曲 + removeSongs 容错 (songloft-org/songloft-plugin-subsonic[#6](https://github.com/songloft-org/songloft/pull/6)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e58358d`](https://github.com/songloft-org/songloft/commit/e58358dba3272432da8c4a55d1a49bdff4da65b4) - Subsonic 远程歌曲格式显示与电台列表修复 + Windows 缓存兼容 *(PR [#231](https://github.com/songloft-org/songloft/pull/231) by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`dd2f902`](https://github.com/songloft-org/songloft/commit/dd2f902965167a1165be66be742fa505be732dc7) - update CHANGELOG for v2.8.10 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`1fd17f5`](https://github.com/songloft-org/songloft/commit/1fd17f50223fe97d6b962e3ba5a3768830cead11) - 在 README 中增加 Kodi 插件客户端描述 *(PR [#232](https://github.com/songloft-org/songloft/pull/232) by [@altman08](https://github.com/altman08))*

### :wrench: Chores
- [`5e56377`](https://github.com/songloft-org/songloft/commit/5e5637771d3079e3fad9a57ed058548683012f9d) - release version 2.9.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.10] - 2026-06-24
### :memo: Documentation Changes
- [`e015dba`](https://github.com/songloft-org/songloft/commit/e015dba42403d68c72a94c2dafebcee132361e77) - update CHANGELOG for v2.8.9 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`717c5eb`](https://github.com/songloft-org/songloft/commit/717c5eb9251e1dd7ec72f08f59f020b98a9be00b) - release version 2.8.10 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.9] - 2026-06-24
### :sparkles: New Features
- [`34a6e4c`](https://github.com/songloft-org/songloft/commit/34a6e4ca1a93d92e9a44d95168e9663a5b832bfd) - remote 歌曲播放时自动提取元数据 *(commit by [@hanxi](https://github.com/hanxi))*
- [`0b19732`](https://github.com/songloft-org/songloft/commit/0b19732fc11b1db116a460b2b406d9b0e6ee447c) - **jsplugin**: implement remaining Bridge API operations *(commit by [@hanxi](https://github.com/hanxi))*
- [`0984246`](https://github.com/songloft-org/songloft/commit/0984246439491810b6a58a862ca123e3df87e877) - **jsplugin**: add yt-dlp music import plugin *(commit by [@hanxi](https://github.com/hanxi))*
- [`15dbb79`](https://github.com/songloft-org/songloft/commit/15dbb79c7ac10677ac7028e9854f3b3efb1c04b9) - support AIF/AIFF format scanning and metadata extraction *(commit by [@hanxi](https://github.com/hanxi))*
- [`15b47b8`](https://github.com/songloft-org/songloft/commit/15b47b8578860d8632330aa84d9b8d3a2ed82b1c) - add AIFF write support and update docs *(commit by [@hanxi](https://github.com/hanxi))*
- [`c9cf78e`](https://github.com/songloft-org/songloft/commit/c9cf78e4cf6bfee0f8647ae5fc32572eb5ddad7e) - **subsonic**: 补全 Subsonic 协议支持，修复音流等客户端兼容性问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`bd98b73`](https://github.com/songloft-org/songloft/commit/bd98b73d4d775ed5a10392bc53300745d6e92623) - **jsplugin**: 新增 plugin.getNetworkAddresses bridge API 并更新 miot 插件 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`96a1718`](https://github.com/songloft-org/songloft/commit/96a171819bbf7ce42a217f66a18e57ca88d5fce7) - 解决 http proxy 问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`9048266`](https://github.com/songloft-org/songloft/commit/90482663f80b5cd2eb2da51ed9ae3c792145589d) - **jsplugin**: 修复禁用插件后状态被自愈机制覆盖回 active 的竞态问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`962950c`](https://github.com/songloft-org/songloft/commit/962950cd716d24f4dc67643898a512c08662d4a5) - **jsplugin**: 电台插件 M3U 导入大小限制从 5MB 提升到 20MB *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`0aaf384`](https://github.com/songloft-org/songloft/commit/0aaf384a4634c9fdf1faa66ef27b65490e7ef291) - update CHANGELOG for v2.8.8 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`f162a8f`](https://github.com/songloft-org/songloft/commit/f162a8f975cf2404294bcb1888c80b6c046dcbac) - update supported formats to include AIF/AIFF *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`e371ff7`](https://github.com/songloft-org/songloft/commit/e371ff7913fcd228e0d34fc18e554d93a1f9ef4a) - update radio plugin submodule (v2026.6.23) *(commit by [@hanxi](https://github.com/hanxi))*
- [`8391804`](https://github.com/songloft-org/songloft/commit/8391804b30853a8c4b67f3c511bb2538f4f83d35) - update songloft-player submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`15a7f33`](https://github.com/songloft-org/songloft/commit/15a7f33ca2c560f57f0d26908b87fd3064494225) - update subsonic plugin submodule (歌单浏览功能) *(commit by [@hanxi](https://github.com/hanxi))*
- [`5b87148`](https://github.com/songloft-org/songloft/commit/5b8714879dfc0b9115a43c470caf48b7528bf31d) - update miot plugin submodule (升级 SDK 2.6.3) *(commit by [@hanxi](https://github.com/hanxi))*
- [`ec153c9`](https://github.com/songloft-org/songloft/commit/ec153c9e2b1e1db2718adafa170e26c4a35ca13d) - update subsonic/dav plugin submodules (icon + v2.1.1/v1.1.1) *(commit by [@hanxi](https://github.com/hanxi))*
- [`aa3a4e8`](https://github.com/songloft-org/songloft/commit/aa3a4e8624448ba1c0ff7f61d1250826366c0b76) - update miot plugin submodule (设置页 UI 补全) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b5b7021`](https://github.com/songloft-org/songloft/commit/b5b7021c2324a6984b159995704ab71ae9cc7ef4) - update miot plugin submodule (外部搜索下拉列表) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b1ae529`](https://github.com/songloft-org/songloft/commit/b1ae5290c4cb36bcd6f91de31c069b9d578dc814) - update miot plugin submodule (外部搜索超时可配置) *(commit by [@hanxi](https://github.com/hanxi))*
- [`f5aef4e`](https://github.com/songloft-org/songloft/commit/f5aef4e2438152b9dfec532cbff16367135a5a84) - release version 2.8.9 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.8] - 2026-06-22
### :sparkles: New Features
- [`6e36dcb`](https://github.com/songloft-org/songloft/commit/6e36dcb377d413653da16c56f3485207e329dbd6) - **jsplugin**: 插件源支持 Bearer Token 认证，用于私有源分发 *(commit by [@hanxi](https://github.com/hanxi))*
- [`90a3f19`](https://github.com/songloft-org/songloft/commit/90a3f194a73f12f859427f64cebfbf583e808c0b) - 封面支持插件相对 URL 解析 + AddRemoteSongs 新增 lyric_remote_url 直传字段 *(PR [#203](https://github.com/songloft-org/songloft/pull/203) by [@hanxi](https://github.com/hanxi))*
- [`9f7eaf7`](https://github.com/songloft-org/songloft/commit/9f7eaf7415460e569e6a35322463b898c1139452) - 添加 songloft-plugin-radio 子模块 + 更新 registry *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`5f0cec3`](https://github.com/songloft-org/songloft/commit/5f0cec349040055f18677c396cb58de1dad6e586) - **models**: LyricURLPath 对 remote 歌曲始终返回歌词端点 URL *(PR [#201](https://github.com/songloft-org/songloft/pull/201) by [@hanxi](https://github.com/hanxi))*
- [`8a73b0a`](https://github.com/songloft-org/songloft/commit/8a73b0a0945cbd52ef585cde2238a57419bb6462) - **services**: 避免 exec.LookPath 使用 faccessat2 导致 Termux 上 SIGSYS 崩溃 *(PR [#202](https://github.com/songloft-org/songloft/pull/202) by [@hanxi](https://github.com/hanxi))*
- [`ce441f2`](https://github.com/songloft-org/songloft/commit/ce441f210b565331adda728539a18b53c73b11ac) - **miot**: 更新子模块引用，修复密码/Token登录 token 续期问题 *(PR [#200](https://github.com/songloft-org/songloft/pull/200) by [@hanxi](https://github.com/hanxi))*
- [`b3e077f`](https://github.com/songloft-org/songloft/commit/b3e077faefe16b4961c49da93a44ddcd84e0f935) - **jsplugin**: 私有源 token 仅对同 host 的 includes 透传，防止跨域泄露 *(commit by [@hanxi](https://github.com/hanxi))*
- [`046cf14`](https://github.com/songloft-org/songloft/commit/046cf144cd9b657925e87c7c89adcb42adf34a85) - 修正开发版下载链接，release tag 应为 dev 而非 main *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`947881b`](https://github.com/songloft-org/songloft/commit/947881b1cef0f045c849b2cd022aa78d73bcf30a) - update CHANGELOG for v2.8.7 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`738adfd`](https://github.com/songloft-org/songloft/commit/738adfd9234e5c4aca16a48c19c9a391f3b7f2c9) - 补充插件源私有认证文档 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ab06610`](https://github.com/songloft-org/songloft/commit/ab066104fcd13ac548fdc776ae3e5d768888a968) - 免责声明新增侵权举报渠道（GitHub Issues + 邮箱） *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`7ab59e3`](https://github.com/songloft-org/songloft/commit/7ab59e3b3ec9f456c57ed0b848ad36a6b2f3a9a5) - **legal**: 补全字体许可证，清理插件子模块，更新文档引用 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ffec8e9`](https://github.com/songloft-org/songloft/commit/ffec8e9458e6872b1e282cd23b253fb3c54adef7) - 更新 songloft-player 子模块引用 *(commit by [@hanxi](https://github.com/hanxi))*
- [`9ee94e0`](https://github.com/songloft-org/songloft/commit/9ee94e097bd7ce7ce5f4c1edf73f5e47f42a45a9) - release version 2.8.8 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.7] - 2026-06-20
### :sparkles: New Features
- [`5e79cdc`](https://github.com/songloft-org/songloft/commit/5e79cdcc41d244d7fa807a6d2497138f083f25b0) - **settings**: 用户偏好跨设备同步 *(PR [#196](https://github.com/songloft-org/songloft/pull/196) by [@hanxi](https://github.com/hanxi))*
- [`66748b0`](https://github.com/songloft-org/songloft/commit/66748b06f57afdeaeaa6d546b3556194a8dcaeaa) - **metadata**: 扩展远程歌曲元数据刷新 *(PR [#195](https://github.com/songloft-org/songloft/pull/195) by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`dcaf549`](https://github.com/songloft-org/songloft/commit/dcaf549bfb4d56106c5f039966e086d9f78f06dd) - **miot**: 播放控制栏底部安全区域适配 *(PR [#192](https://github.com/songloft-org/songloft/pull/192) by [@hanxi](https://github.com/hanxi))*
- [`a135f18`](https://github.com/songloft-org/songloft/commit/a135f18da5a0ba2026d61e3744e65bbbbc2f59f1) - **miot**: 修复 Web 端停止播放后进度条持续走动 *(PR [#191](https://github.com/songloft-org/songloft/pull/191) by [@hanxi](https://github.com/hanxi))*
- [`7d00a65`](https://github.com/songloft-org/songloft/commit/7d00a655ed0c875ad08311fa863149df1180aff9) - **miot**: 修复语音口令含歌手名时误匹配其他歌曲 *(PR [#199](https://github.com/songloft-org/songloft/pull/199) by [@hanxi](https://github.com/hanxi))*
- [`37860bd`](https://github.com/songloft-org/songloft/commit/37860bd6c54f4ee00e7566ddeb372fb536177b71) - **miot**: 语音口令歌单匹配优化 *(PR [#198](https://github.com/songloft-org/songloft/pull/198) by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`f65c86b`](https://github.com/songloft-org/songloft/commit/f65c86b004f5ba549077985c73f6fd619373ef30) - update CHANGELOG for v2.8.6 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`d8d9f0b`](https://github.com/songloft-org/songloft/commit/d8d9f0b9c41aaa2416dbdb6c08c43f5e4ba81623) - release version 2.8.7 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.6] - 2026-06-19
### :sparkles: New Features
- [`2341268`](https://github.com/songloft-org/songloft/commit/23412687dc4050de1e53146a69b2f0bcfe6fcb87) - **songs**: 批量刷新远程歌曲时长 API *(PR [#185](https://github.com/songloft-org/songloft/pull/185) by [@hanxi](https://github.com/hanxi))*
- [`2104a03`](https://github.com/songloft-org/songloft/commit/2104a0377741175006b0853f475eb373ca4969bb) - **a11y**: 全面改进无障碍支持 *(PR [#186](https://github.com/songloft-org/songloft/pull/186) by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`6db9455`](https://github.com/songloft-org/songloft/commit/6db945554f0808436021ffa6f19c88852286da77) - **player**: 随机播放模式下全部播放不再固定从第一首开始 *(PR [#184](https://github.com/songloft-org/songloft/pull/184) by [@hanxi](https://github.com/hanxi))*
- [`6c9590a`](https://github.com/songloft-org/songloft/commit/6c9590a881dd70f81ade84a8e47a2bbb38c94a69) - **cache**: 流式播放缓存后自动回填歌曲时长 *(PR [#185](https://github.com/songloft-org/songloft/pull/185) by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`d036d02`](https://github.com/songloft-org/songloft/commit/d036d02a62ee370af9bf7bbe1fecbb3eb49a9404) - update CHANGELOG for v2.8.5 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`50e9ce9`](https://github.com/songloft-org/songloft/commit/50e9ce94c40d8a4419237de8753dfc3967d66c60) - update songloft-player submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`925a2a7`](https://github.com/songloft-org/songloft/commit/925a2a7ef100c633645ce295b76448ebd2e0a87a) - update songloft-plugin-miot submodule *(commit by [@hanxi](https://github.com/hanxi))*
- [`15ebc80`](https://github.com/songloft-org/songloft/commit/15ebc802b1d1c395c39fbc5f52ea28a5709b3505) - release version 2.8.6 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.5] - 2026-06-16
### :sparkles: New Features
- [`5744cc3`](https://github.com/songloft-org/songloft/commit/5744cc35db02d0d0c53e6f1ad92746ced61fa73b) - **nav**: 放宽 tab 数量限制至 10，移动端支持「更多」溢出菜单 *(commit by [@hanxi](https://github.com/hanxi))*
- [`389be4a`](https://github.com/songloft-org/songloft/commit/389be4a4900554dd1cef5a460a09e8058332a0c0) - **playlist**: 更新 songloft-player 子模块，歌单列表高亮当前播放歌单 (close [#182](https://github.com/songloft-org/songloft/pull/182)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`3c37727`](https://github.com/songloft-org/songloft/commit/3c3772728fd3aeb4a96f3bf886628279c1c99218) - **plugin**: 新增歌词提供者回调机制和 lrclib 歌词插件 (close [#183](https://github.com/songloft-org/songloft/pull/183)) *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`bec658c`](https://github.com/songloft-org/songloft/commit/bec658cbafde9b921cee8e6f4daa59f2a4f9c6d9) - **web**: 更新 songloft-player 子模块，修复 embedded 模式字体缺失 (close [#177](https://github.com/songloft-org/songloft/pull/177)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e1b7d40`](https://github.com/songloft-org/songloft/commit/e1b7d408e44c47a83b5df116d8055cc66d24a9ea) - **playlist**: 修复从歌曲选择封面后预览空白及缓存不刷新 (close [#176](https://github.com/songloft-org/songloft/pull/176)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`4c01988`](https://github.com/songloft-org/songloft/commit/4c01988f47a5bf50fbbb4c732c3262f45bb70772) - **plugin**: 更新 songloft-player 子模块，修复 Windows 最小化后插件 WebView 拦截桌面右键 (close [#181](https://github.com/songloft-org/songloft/pull/181)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`1eb05ac`](https://github.com/songloft-org/songloft/commit/1eb05ac89536efa2a3771719f929281193e4353f) - **plugin**: 修复 DAV 插件 buildStreamUrl 路径双重前缀导致播放 404 (close [#180](https://github.com/songloft-org/songloft/pull/180)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`d5ec3cc`](https://github.com/songloft-org/songloft/commit/d5ec3cc3b9b0678d783541eeee81911a01472f23) - **plugin**: URLSearchParams polyfill 支持对象参数，修复歌词插件精确搜索参数丢失 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`690c6cc`](https://github.com/songloft-org/songloft/commit/690c6ccdb40dc41596ed8697eeeb62248419542a) - update CHANGELOG for v2.8.4 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`12b565d`](https://github.com/songloft-org/songloft/commit/12b565d45e9c6f8706df6ca8a5b4215fe2739668) - **plugin**: 更新歌词插件子模块，补全元数据和发布流程 *(commit by [@hanxi](https://github.com/hanxi))*
- [`841fffc`](https://github.com/songloft-org/songloft/commit/841fffc0af44fb332fdd6c97d7b044da16b1af2a) - **plugin**: 更新插件源子模块，添加歌词搜索插件 *(commit by [@hanxi](https://github.com/hanxi))*
- [`27be42d`](https://github.com/songloft-org/songloft/commit/27be42d92c92c0d3f7aaaad88aac4e599577ef22) - release version 2.8.5 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.4] - 2026-06-15
### :bug: Bug Fixes
- [`422c968`](https://github.com/songloft-org/songloft/commit/422c968f60fae1a70bd50bc09334b848148d9bff) - **miot**: 更新 miot 插件 - 修复语音搜歌匹配错误 (close [#83](https://github.com/songloft-org/songloft/pull/83)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e127621`](https://github.com/songloft-org/songloft/commit/e127621150d023390c44fd6aa282b077ba23d121) - **playlist**: 修复编辑歌单从歌曲选择封面无效的问题 (close [#176](https://github.com/songloft-org/songloft/pull/176)) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`ca6eae8`](https://github.com/songloft-org/songloft/commit/ca6eae8c425c95a40f5806d9142a5439576812f8) - update CHANGELOG for v2.8.3 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`379bf51`](https://github.com/songloft-org/songloft/commit/379bf51d2c761cadb5aa32c38fe3df2a3301a322) - release version 2.8.4 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.3] - 2026-06-15
### :sparkles: New Features
- [`00763f2`](https://github.com/songloft-org/songloft/commit/00763f26c7096888d911287d23f9d5842e9bdd52) - 下载歌曲时拉取 URL 歌词写入文件，实现 MP4/OGG 元数据写入 *(commit by [@hanxi](https://github.com/hanxi))*
- [`7f98bae`](https://github.com/songloft-org/songloft/commit/7f98bae3f7ef7bc921abe17a402156749f8c75d2) - APE 封面读写支持，更新 tag 写入文档 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`80d6a8f`](https://github.com/songloft-org/songloft/commit/80d6a8fe48d0764f08028819e81dfa635c63d015) - **player**: 低码率音质下播放按钮需按两次才能播放 (close [#170](https://github.com/songloft-org/songloft/pull/170)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`fc03e78`](https://github.com/songloft-org/songloft/commit/fc03e78fdbef5b8f50002846ba9b6ca0c3fa3b03) - **player**: 所有播放来源广播 onPlayEvent 事件 (close [#173](https://github.com/songloft-org/songloft/pull/173)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`69554eb`](https://github.com/songloft-org/songloft/commit/69554eb7c59168b6b46024fb9fb07c48631bc6fb) - **player**: 歌词加载后立即推送到灵动岛 (close [#98](https://github.com/songloft-org/songloft/pull/98)) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`8cf69a9`](https://github.com/songloft-org/songloft/commit/8cf69a9fbd9b7ecf5b428d45fa3dcfe9f1dee477) - update CHANGELOG for v2.8.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`e7d1341`](https://github.com/songloft-org/songloft/commit/e7d13412039d6b5a559f83daad46565d7440c3bd) - 添加 Issue 模板和模板选择器配置 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f2e6777`](https://github.com/songloft-org/songloft/commit/f2e677767ed391f35911e9cb7a4c9e3e3e36536a) - 移除网络歌曲转本地功能宣传，downloader 插件定位为示例 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`0fc38f5`](https://github.com/songloft-org/songloft/commit/0fc38f53fb7b5296475eb8391ac0322bee03d408) - release version 2.8.3 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.2] - 2026-06-13
### :sparkles: New Features
- [`feb75bb`](https://github.com/songloft-org/songloft/commit/feb75bb142b1ef2ec5c0e4d5f822b392a5c3f060) - **jsplugin**: add onPlayEvent callback for play event subscription (close [#164](https://github.com/songloft-org/songloft/pull/164)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`23311d3`](https://github.com/songloft-org/songloft/commit/23311d3007bcd88477247bd838ee680de4d97a66) - **transcode**: 支持多音质转码与按码率缓存 (close [#169](https://github.com/songloft-org/songloft/pull/169)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`7ea30d2`](https://github.com/songloft-org/songloft/commit/7ea30d2fc9e85c9a8cf0b8926b00ebb68f4c6ff1) - **playlist**: 歌单歌曲支持排序与搜索 (close [#168](https://github.com/songloft-org/songloft/pull/168)) *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`0737849`](https://github.com/songloft-org/songloft/commit/0737849e273db2ae046696c71b63d9b1ec9eeca5) - **miot**: 修复语音调音量"百分之X"被误解析为100% (close [#166](https://github.com/songloft-org/songloft/pull/166)) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e31f67d`](https://github.com/songloft-org/songloft/commit/e31f67d1b44b003e68ff0d4ddae3e2b142268894) - **jsruntime**: HealthProbe 污染 eval 超时导致定时器驱动操作被误中断 (close songloft-org/songloft-plugin-miot[#10](https://github.com/songloft-org/songloft/pull/10)) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`136fbf3`](https://github.com/songloft-org/songloft/commit/136fbf388605c24ba4d7030e02e6759e63994712) - update CHANGELOG for v2.8.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`0bac21d`](https://github.com/songloft-org/songloft/commit/0bac21de3de3adffdf20a81c494bbe87e8a37250) - **submodule**: update jsplugins and player submodules *(commit by [@hanxi](https://github.com/hanxi))*
- [`cecabad`](https://github.com/songloft-org/songloft/commit/cecabadcf7cce6bbe405c7e597b1b72682e9ae78) - **submodule**: update player and plugin-toolchain for play event support *(PR [#164](https://github.com/songloft-org/songloft/pull/164) by [@hanxi](https://github.com/hanxi))*
- [`65d806d`](https://github.com/songloft-org/songloft/commit/65d806d274590bce118355d6f14c1119915ee7e7) - release version 2.8.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.1] - 2026-06-12
### :sparkles: New Features
- [`9705163`](https://github.com/songloft-org/songloft/commit/970516327fa8fcbeee7939ee8e04fbde063541af) - **cache**: remove convert-to-local feature and add custom cache directory *(commit by [@hanxi](https://github.com/hanxi))*
- [`6c35b39`](https://github.com/songloft-org/songloft/commit/6c35b3909ca5117ed2b483e2b1aece6c810a7f9c) - **cache**: streaming proxy with cache_path + song download plugin *(commit by [@hanxi](https://github.com/hanxi))*
- [`2a1ef4e`](https://github.com/songloft-org/songloft/commit/2a1ef4ebf30f272d7890022ec0044486f7ea24d9) - **model**: 添加 source_cover_url 字段并统一编辑表单短域显示 *(commit by [@hanxi](https://github.com/hanxi))*
- [`86d1805`](https://github.com/songloft-org/songloft/commit/86d1805efb5969d5b6dc7a712d35350a47bd6d34) - **jsplugin**: 新安装插件默认启用 *(commit by [@hanxi](https://github.com/hanxi))*
- [`de30985`](https://github.com/songloft-org/songloft/commit/de30985d140096e9d84b1e686278f23c146d5d7c) - 新增自动创建歌单排除目录 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`821aa41`](https://github.com/songloft-org/songloft/commit/821aa4107b7a01c62cfdb0c44ce0ce72ffd771b0) - **jsplugin**: refresh publicPaths at runtime without restart *(PR [#158](https://github.com/songloft-org/songloft/pull/158) by [@hanxi](https://github.com/hanxi))*
- [`f3a28bc`](https://github.com/songloft-org/songloft/commit/f3a28bcd9bb99f3cae176779e9f9090536d5abb6) - **fingerprint**: remove invalid length threshold in ExtractFingerprint
- [`bd2aec9`](https://github.com/songloft-org/songloft/commit/bd2aec91cc1e714b25a946f939e38e37a66fd5be) - **cache**: 修复 Windows 编译失败 — syscall.Statfs 不跨平台 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`b023c27`](https://github.com/songloft-org/songloft/commit/b023c2729beacb7a7f197cf9902dabfa11a92a42) - update CHANGELOG for v2.8.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`dc661e8`](https://github.com/songloft-org/songloft/commit/dc661e82f45fb63e519cb5c269bad223573ec560) - **repowiki**: regenerate wiki from latest codebase *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`8bb65ea`](https://github.com/songloft-org/songloft/commit/8bb65eae54e61568efef05368c65a4683c6d9b1b) - **miot**: update submodule ref for [#157](https://github.com/songloft-org/songloft/pull/157) fix *(commit by [@hanxi](https://github.com/hanxi))*
- [`95dd299`](https://github.com/songloft-org/songloft/commit/95dd2997c87a3bb7636f0ae8f63a00bbbe09dc2f) - **miot**: update submodule ref for [#155](https://github.com/songloft-org/songloft/pull/155) indicator light fix *(commit by [@hanxi](https://github.com/hanxi))*
- [`2396230`](https://github.com/songloft-org/songloft/commit/239623054fa2b20a6895027a5afc7d3024170fa4) - **submodule**: update songloft-plugin-downloader *(commit by [@hanxi](https://github.com/hanxi))*
- [`83555f4`](https://github.com/songloft-org/songloft/commit/83555f492ccacd5be699f304753678bc5764ed15) - **submodule**: update songloft-plugin-downloader *(commit by [@hanxi](https://github.com/hanxi))*
- [`49eb151`](https://github.com/songloft-org/songloft/commit/49eb1514424e906cc1699c2e7bd7d70cc349d42f) - **submodule**: update jsplugins package-lock.json *(commit by [@hanxi](https://github.com/hanxi))*
- [`6213546`](https://github.com/songloft-org/songloft/commit/6213546720877f11ad2f177a213a6aacdd8ded00) - release version 2.8.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.8.0] - 2026-06-11
### :sparkles: New Features
- [`da56b08`](https://github.com/songloft-org/songloft/commit/da56b087f3d63172931ba26137099e5b70ce6b71) - **scan**: add toggle to enable/disable auto-create playlists on scan *(commit by [@hanxi](https://github.com/hanxi))*
- [`f2a6c1e`](https://github.com/songloft-org/songloft/commit/f2a6c1e26c00592e3b0e7352e025218bdf607b48) - **jsplugin**: externalPaths 改为 manifest 声明，删除管理员 API *(PR [#151](https://github.com/songloft-org/songloft/pull/151) by [@hanxi](https://github.com/hanxi))*
- [`f313243`](https://github.com/songloft-org/songloft/commit/f313243fe0846459a77e9f08169214196e08756e) - **api**: 暴露 lyric_remote_url 字段，支持编辑网络歌曲歌词 URL *(PR [#141](https://github.com/songloft-org/songloft/pull/141) by [@hanxi](https://github.com/hanxi))*
- [`afc45a6`](https://github.com/songloft-org/songloft/commit/afc45a63b024928514f874736bae2cb0b911b886) - **tracely**: 支持安装/升级统计上报，配置改为三变量注入 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`99a4d4f`](https://github.com/songloft-org/songloft/commit/99a4d4f704784feb09cdb7db7e9c7fe6df4f0aef) - [#145](https://github.com/songloft-org/songloft/pull/145) [#147](https://github.com/songloft-org/songloft/pull/147) - WriteTags damaged cover cleanup + playlist cover fallback to songs *(PR [#150](https://github.com/songloft-org/songloft/pull/150) by [@laihya](https://github.com/laihya))*
- [`1e76796`](https://github.com/songloft-org/songloft/commit/1e767963e309daf0882d4150bc6ddfffdd997a4c) - **jsplugin**: resolveFSPath 支持 music:// 和绝对路径 *(PR [#151](https://github.com/songloft-org/songloft/pull/151) by [@hanxi](https://github.com/hanxi))*
- [`d0c7140`](https://github.com/songloft-org/songloft/commit/d0c71404620fc365622780818ef01d07eb348354) - **scan**: chromaprint 指纹被歌词/元数据污染 *(PR [#146](https://github.com/songloft-org/songloft/pull/146) by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`e43a1c9`](https://github.com/songloft-org/songloft/commit/e43a1c917221ba5eb822a9de99c62b192f03e868) - update CHANGELOG for v2.7.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`a2c4860`](https://github.com/songloft-org/songloft/commit/a2c4860b0b9f31a66b78d8c0a90f13e5ae91e76b) - bump plugin submodules (tag v1.0.6, dav v1.0.4, subsonic v2.0.1) *(commit by [@hanxi](https://github.com/hanxi))*
- [`dceab9e`](https://github.com/songloft-org/songloft/commit/dceab9e397e4d3b140e073b944caa90054bb0785) - release version 2.8.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.7.0] - 2026-06-09
### :sparkles: New Features
- [`9b3fb55`](https://github.com/songloft-org/songloft/commit/9b3fb55cd59e62f74eefefb1466e0634c549d147) - **jsplugin**: support pathPrefix param in songs.list bridge *(commit by [@hanxi](https://github.com/hanxi))*
- [`59f8d0c`](https://github.com/songloft-org/songloft/commit/59f8d0c1cff80bc808b4550d354fbf292c471d10) - **song**: add ISRC field to Song model and extract from audio tags *(commit by [@hanxi](https://github.com/hanxi))*
- [`4ccbc0c`](https://github.com/songloft-org/songloft/commit/4ccbc0ccb58eabcfc025c1e526e27794e320f9b4) - **scan**: add auto-scan with file stability detection *(commit by [@hanxi](https://github.com/hanxi))*
- [`24f934d`](https://github.com/songloft-org/songloft/commit/24f934d5bb45995312c835f857d66a3f0bf51a45) - 核心镜像缺ALSA用户态运行文件 *(PR [#135](https://github.com/songloft-org/songloft/pull/135) by [@huaimi123](https://github.com/huaimi123))*
- [`5b8d062`](https://github.com/songloft-org/songloft/commit/5b8d0621d0f3b6387492b69c2a336742092ab0fc) - **jsplugin**: add serveFile directive, file serve route, and publicPaths *(commit by [@hanxi](https://github.com/hanxi))*
- [`139b667`](https://github.com/songloft-org/songloft/commit/139b667bcef8edf43ffe8c628580a11514e85d36) - **jsplugin**: add external-paths settings API *(commit by [@hanxi](https://github.com/hanxi))*
- [`e7c365c`](https://github.com/songloft-org/songloft/commit/e7c365ca2de450ace76d3ee5da498617458e899b) - **subsonic**: add Subsonic server mode *(commit by [@hanxi](https://github.com/hanxi))*
- [`b5ebe57`](https://github.com/songloft-org/songloft/commit/b5ebe572138212b2ab4047f482ca40993c951895) - **subsonic**: add server mode settings UI *(commit by [@hanxi](https://github.com/hanxi))*
- [`afd326e`](https://github.com/songloft-org/songloft/commit/afd326ea3fdb00882c29c6957bebc5e0887a0acf) - **subsonic**: add getGenres endpoint *(commit by [@hanxi](https://github.com/hanxi))*
- [`51eab49`](https://github.com/songloft-org/songloft/commit/51eab497becda990a1ab786ad9a7b0b242298e63) - **subsonic**: add getSong, getStarred, getIndexes endpoints *(commit by [@hanxi](https://github.com/hanxi))*
- [`383b6c1`](https://github.com/songloft-org/songloft/commit/383b6c1a7f59c7d197f1af5d9bf5d1259920b151) - **subsonic**: getStarred reads from built-in favorites playlist *(commit by [@hanxi](https://github.com/hanxi))*
- [`f3f9b85`](https://github.com/songloft-org/songloft/commit/f3f9b85083ba1e0f05527909a98ba4fa90d2d7a3) - **subsonic**: add copy button for server URL *(commit by [@hanxi](https://github.com/hanxi))*
- [`fdddf45`](https://github.com/songloft-org/songloft/commit/fdddf45c0c4b6e4e536336cb3932cb4f65c33723) - **jsplugin**: add plugin icon support *(PR [#139](https://github.com/songloft-org/songloft/pull/139) by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`fc52d35`](https://github.com/songloft-org/songloft/commit/fc52d355163d1e28f9df799c77d1fdb5b7993a76) - **miot**: embed模式下搜索框被遮挡 *(commit by [@hanxi](https://github.com/hanxi))*
- [`4e06ba5`](https://github.com/songloft-org/songloft/commit/4e06ba5e35b9af087a76404d64dad2bb206b4914) - **plugin-toolchain**: add fs:music, fs:external to builder permission whitelist *(commit by [@hanxi](https://github.com/hanxi))*
- [`59e2f65`](https://github.com/songloft-org/songloft/commit/59e2f65668644a8d01743046cfecd2753dcf2baf) - **jsplugin**: publicPaths bypass via AuthMiddleware checker *(commit by [@hanxi](https://github.com/hanxi))*
- [`1997a90`](https://github.com/songloft-org/songloft/commit/1997a900be041c22abfe41b34e356bbe230c3f18) - **subsonic**: server URL display should not include /rest suffix *(commit by [@hanxi](https://github.com/hanxi))*
- [`48d4939`](https://github.com/songloft-org/songloft/commit/48d4939847c1bcdc25ae3110db931d2e3812cfac) - **subsonic**: fix field mapping and add getAlbum/getArtist endpoints *(commit by [@hanxi](https://github.com/hanxi))*
- [`a927818`](https://github.com/songloft-org/songloft/commit/a927818006cb2f3356205094de347b9c3a3996e7) - **subsonic**: add required song fields for client compatibility *(commit by [@hanxi](https://github.com/hanxi))*
- [`66a0114`](https://github.com/songloft-org/songloft/commit/66a0114bc5a741a4c1f4aab4d77068bb5e47d91c) - **subsonic**: fix cover art and add lyrics support *(commit by [@hanxi](https://github.com/hanxi))*
- [`e26f71e`](https://github.com/songloft-org/songloft/commit/e26f71e690cc850245a6660386ba0f8473c6b049) - **subsonic**: playlist cover art display *(commit by [@hanxi](https://github.com/hanxi))*

### :recycle: Refactors
- [`d7a0d89`](https://github.com/songloft-org/songloft/commit/d7a0d8902af160535b7ea24c2349e29bb1680a92) - **jsplugin**: merge manifest.json into plugin.json *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`d5810c5`](https://github.com/songloft-org/songloft/commit/d5810c5cf6afbffd86e2cbf57923acafe6772daa) - update CHANGELOG for v2.6.4 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`eda34ed`](https://github.com/songloft-org/songloft/commit/eda34eda75a703d6c5e842c6895de54a7bf5c43e) - **subsonic**: rename plugin title *(commit by [@hanxi](https://github.com/hanxi))*
- [`59b7541`](https://github.com/songloft-org/songloft/commit/59b754127a7ccf5c5734133a2b5432a2b01d6841) - release version 2.7.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.6.4] - 2026-06-07
### :sparkles: New Features
- [`27aeb03`](https://github.com/songloft-org/songloft/commit/27aeb035197f23454c3d6f39772dd6d13edf8871) - **jsplugin**: 支持插件强制更新（跳过版本检查） *(commit by [@hanxi](https://github.com/hanxi))*
- [`0cce183`](https://github.com/songloft-org/songloft/commit/0cce183868bdd5a916e4ed0f1a94db767f6af57b) - **fingerprint**: 支持重新计算全部音频指纹 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`d617b20`](https://github.com/songloft-org/songloft/commit/d617b20efacc85d6d7c3abd09eedc80cc0e04b05) - 修复 WAV/APE 标签及文件名的中文编码乱码问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`f63c93a`](https://github.com/songloft-org/songloft/commit/f63c93aa6ad8a3765eada72f6d6d31baa0d414b0) - update CHANGELOG for v2.6.3 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`cd1a65f`](https://github.com/songloft-org/songloft/commit/cd1a65f5fef294d4f2648a4428bbf5d6b6485715) - release version 2.6.4 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.6.3] - 2026-06-07
### :sparkles: New Features
- [`5dbbff9`](https://github.com/songloft-org/songloft/commit/5dbbff9f8833b0d44c0fe4989f9d56cbc7a5401d) - 用 ffmpeg chromaprint 替代 fpcalc，扫描后自动计算指纹 *(commit by [@hanxi](https://github.com/hanxi))*
- [`fe7810c`](https://github.com/songloft-org/songloft/commit/fe7810cb7d1cb369b6636e8fc94b7ef6afd789c6) - 更新 miot 插件，支持定时开关对话监听 *(commit by [@hanxi](https://github.com/hanxi))*
- [`4fbfdcb`](https://github.com/songloft-org/songloft/commit/4fbfdcbb4ec19b25ae3e38d026c0b76ddf07194b) - 重构插件主题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ca52f4e`](https://github.com/songloft-org/songloft/commit/ca52f4eb3783d6fea92a0f7432d37c438cb37501) - **dedup**: 指纹去重删除时同步删除音频文件 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`cec5a8a`](https://github.com/songloft-org/songloft/commit/cec5a8ae25cff8d1bfdb44ef68ff9afbbd953609) - 设置 PKG_CONFIG_PATH 让 ffmpeg configure 找到 libchromaprint *(commit by [@hanxi](https://github.com/hanxi))*
- [`d4a8cc4`](https://github.com/songloft-org/songloft/commit/d4a8cc40722029fce6210cfb52766585b45970c1) - 设置 PKG_CONFIG_PATH 让 ffmpeg configure 找到 libchromaprint *(commit by [@hanxi](https://github.com/hanxi))*
- [`872abaf`](https://github.com/songloft-org/songloft/commit/872abaf8546c8795ee85e46c7252ff43e670380b) - **tag**: 移除与 common.css 重复的导航样式，修复非 embed 模式 tab 异常 *(commit by [@hanxi](https://github.com/hanxi))*
- [`8336676`](https://github.com/songloft-org/songloft/commit/8336676e53e69043fbe8c23c084685272d53e456) - **tag**: 移除与 common.css 重复的 card 样式定义 *(commit by [@hanxi](https://github.com/hanxi))*
- [`9c8db18`](https://github.com/songloft-org/songloft/commit/9c8db18c07b1b17da9d8e4b33268c7270fd458c8) - **settings**: 缓存管理操作区域默认折叠，防止误触 *(commit by [@hanxi](https://github.com/hanxi))*

### :zap: Performance Improvements
- [`e8c4d5a`](https://github.com/songloft-org/songloft/commit/e8c4d5afd5ef6787b4368fd696ad378ac9a9fb0c) - 写入标签前预检，标签一致时跳过磁盘写入 *(commit by [@hanxi](https://github.com/hanxi))*

### :recycle: Refactors
- [`a4459ad`](https://github.com/songloft-org/songloft/commit/a4459ad44560ac2207892881c75c9f1ddd4e5c8a) - **tag**: 导航改为底部 tab-bar，统一插件 UI 风格 *(commit by [@hanxi](https://github.com/hanxi))*
- [`87c217f`](https://github.com/songloft-org/songloft/commit/87c217ff2f51e484af8c2bffab93133c83823e18) - **tag-plugin**: 移除 fpcalc 依赖，改用主程序指纹 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`9882bb0`](https://github.com/songloft-org/songloft/commit/9882bb099ae1951e382ec4a53a4829b6f631feb5) - update CHANGELOG for v2.6.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*
- [`bd688b2`](https://github.com/songloft-org/songloft/commit/bd688b2b77e69a924cab380dcb17713dc483977c) - **faq**: 添加多音乐目录配置方法 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`316bbd6`](https://github.com/songloft-org/songloft/commit/316bbd6d9da31d793d692deb48f7696be477e1a5) - 更新 dav/subsonic 子模块，清理死代码 *(commit by [@hanxi](https://github.com/hanxi))*
- [`28da741`](https://github.com/songloft-org/songloft/commit/28da74119db465c8c16567510a8290b9d0031699) - 更新 miot 子模块，搜索前打断音箱播报 *(commit by [@hanxi](https://github.com/hanxi))*
- [`3306c2c`](https://github.com/songloft-org/songloft/commit/3306c2c6551e20473d9b72a2af28d99b1d34982f) - release version 2.6.3 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.6.2] - 2026-06-06
### :sparkles: New Features
- [`aa9c260`](https://github.com/songloft-org/songloft/commit/aa9c260da96cc844ce858432e1d1e4163837867d) - 新增歌曲去重功能 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`1e9722f`](https://github.com/songloft-org/songloft/commit/1e9722f96a03a6122c9ea10bd8afd7b01daa3ac7) - 更新 pkg/tag 子模块，修复 APE/WAV 读写问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`0b915b6`](https://github.com/songloft-org/songloft/commit/0b915b625b11c1022eab38110ce49bf29bf33192) - release version 2.6.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.6.0] - 2026-06-05
### :sparkles: New Features
- [`7bfe806`](https://github.com/songloft-org/songloft/commit/7bfe80604fa49b67202b5ddecd002cf9566dadca) - 扫描歌曲支持用文件名作为标题 songloft-org/songloft[#100](https://github.com/songloft-org/songloft/pull/100) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b140863`](https://github.com/songloft-org/songloft/commit/b140863aacde2a8ee6c8cada68b1a30ff0f6a7b0) - 实现自定义tab songloft-org/songloft[#103](https://github.com/songloft-org/songloft/pull/103) *(commit by [@hanxi](https://github.com/hanxi))*
- [`e9419a8`](https://github.com/songloft-org/songloft/commit/e9419a8d757c100f963a0005a80bb94780217bc4) - 插件新增 websocket 接口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`d433bd2`](https://github.com/songloft-org/songloft/commit/d433bd2145e878f6311fa8c6c2c2de434761f5e7) - 插件适配嵌入到主程序 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`4cf8a77`](https://github.com/songloft-org/songloft/commit/4cf8a771b9da95fc5ea11a012460bbda62402156) - 修复插件无法运行的问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`2db9ac8`](https://github.com/songloft-org/songloft/commit/2db9ac8de8d9b9d7cd7700a4bd4d9282a5124eff) - 统一插件打包脚本 *(commit by [@hanxi](https://github.com/hanxi))*
- [`236ec5b`](https://github.com/songloft-org/songloft/commit/236ec5b4cf69f519d7b631fb9fc8950ae16eafee) - **convert**: 接入全局 HTTP 代理，修复代理环境下歌曲/封面下载失败 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`d387aaa`](https://github.com/songloft-org/songloft/commit/d387aaa2ce53d31292e04b325dd3167fb6dedaf6) - update CHANGELOG for v2.5.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`93d8f56`](https://github.com/songloft-org/songloft/commit/93d8f5657d5aa4d014e97d17c3492261ba4614cb) - 更新插件子模块及构建脚本 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ae2b657`](https://github.com/songloft-org/songloft/commit/ae2b6577b4cad1012a4089283c6149a75795352f) - 更新插件子模块引用 *(commit by [@hanxi](https://github.com/hanxi))*
- [`5e52336`](https://github.com/songloft-org/songloft/commit/5e523363c03e14332e57ffb5eea29900187c899a) - release version 2.6.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.5.1] - 2026-06-04
### :sparkles: New Features
- [`cba096f`](https://github.com/songloft-org/songloft/commit/cba096f483b7d19295479d41f257bca99fa459bf) - 支持全局HTTP代理 *(commit by [@hanxi](https://github.com/hanxi))*
- [`a0fd898`](https://github.com/songloft-org/songloft/commit/a0fd8983f5920d4f9ee5b64b8e131d0113f03f35) - 新增标签内容写入接口和歌曲文件整理接口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`eabba1d`](https://github.com/songloft-org/songloft/commit/eabba1d6289e34f5a42de8e04efdf4821411b7a6) - 插件新增 fs 接口 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`1c48c37`](https://github.com/songloft-org/songloft/commit/1c48c37645c2aac9c14f28310e99fb6dde126e73) - 部分歌曲标签信息识别不出来 [#95](https://github.com/songloft-org/songloft/pull/95) *(commit by [@hanxi](https://github.com/hanxi))*
- [`adc0019`](https://github.com/songloft-org/songloft/commit/adc0019410521dcef30b71a51524115acfe806c2) - 修复编码问题 songloft-org/songloft[#104](https://github.com/songloft-org/songloft/pull/104) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`58a4ffd`](https://github.com/songloft-org/songloft/commit/58a4ffdd1b39283bc3cc69bf26bce0fa0eea6021) - update CHANGELOG for v2.5.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`41c1528`](https://github.com/songloft-org/songloft/commit/41c15281364d842c20ba0c196580f7e5048080e6) - release version 2.5.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [Unreleased]
### :sparkles: New Features
- 插件商店支持自定义代理输入（与插件更新/系统升级对话框一致的 RadioGroup + 自定义输入框）
- 新增通用 HTTP 代理设置（`/settings/http-proxy`），所有后端外发请求可通过用户配置的代理转发

## [v2.5.0] - 2026-06-04
### :sparkles: New Features
- [`8cd86bb`](https://github.com/songloft-org/songloft/commit/8cd86bb8657fabcc435004dc1185fb1ae8748a39) - 支持一键更新所有插件 songloft-org/songloft[#61](https://github.com/songloft-org/songloft/pull/61) *(commit by [@hanxi](https://github.com/hanxi))*
- [`bda6f9c`](https://github.com/songloft-org/songloft/commit/bda6f9c25a3a3df6a8c656d1860bad2c84eb469a) - 新增插件下载和执行命令 [#90](https://github.com/songloft-org/songloft/pull/90) *(commit by [@hanxi](https://github.com/hanxi))*
- [`22a7270`](https://github.com/songloft-org/songloft/commit/22a7270d7f4068cb9f6bf7c37ef37f1a2eb05d39) - 新增插件源 songloft-org/songloft[#89](https://github.com/songloft-org/songloft/pull/89) *(commit by [@hanxi](https://github.com/hanxi))*
- [`635787a`](https://github.com/songloft-org/songloft/commit/635787a57653d9936f33f4905f8fdfe9daf69ba5) - 新增官方插件源地址 *(commit by [@hanxi](https://github.com/hanxi))*
- [`4f01f2c`](https://github.com/songloft-org/songloft/commit/4f01f2cf38f8580927bcc59d1e661bc90165e039) - 优化版本升级逻辑 *(commit by [@hanxi](https://github.com/hanxi))*
- [`7bd92c5`](https://github.com/songloft-org/songloft/commit/7bd92c5eed70baca9d2c961280bfb87d5194d9fb) - 继续优化首次加载页面速度 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`1596210`](https://github.com/songloft-org/songloft/commit/159621015f49197b3b0972009f846a5712661f6a) - 优化首次加载 songloft-org/songloft[#91](https://github.com/songloft-org/songloft/pull/91) *(commit by [@hanxi](https://github.com/hanxi))*
- [`2d85024`](https://github.com/songloft-org/songloft/commit/2d85024909b1c30bf4981790df7169c7a9dac559) - **http**: 修复 URL userinfo 丢失及 JS URL polyfill 相对路径解析缺陷 *(commit by [@hanxi](https://github.com/hanxi))*
- [`f268fe3`](https://github.com/songloft-org/songloft/commit/f268fe3d49141216ba7397447519a31f430ccfa6) - **proxy**: add basic auth handling and update User-Agent for remote resources *(PR [#93](https://github.com/songloft-org/songloft/pull/93) by [@Dev-Wiki](https://github.com/Dev-Wiki))*
- [`38174a4`](https://github.com/songloft-org/songloft/commit/38174a4d8468b859968b258572ba9a7ddcc741e5) - 修复打包镜像报错 *(commit by [@hanxi](https://github.com/hanxi))*
- [`5e971b3`](https://github.com/songloft-org/songloft/commit/5e971b3586f56b07357fc3721a935a3fbf1705af) - **scan**: 检测并跳过垃圾 tag，回退使用文件名作为标题 *(PR [#94](https://github.com/songloft-org/songloft/pull/94) by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`c3cde9f`](https://github.com/songloft-org/songloft/commit/c3cde9f51f4f352abf65337075a9a10d0e2e2b3e) - update CHANGELOG for v2.4.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`9ad00ad`](https://github.com/songloft-org/songloft/commit/9ad00adbf0226163624f76944740ce4cdd6e72ef) - release version 2.5.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.4.0] - 2026-06-02
### :sparkles: New Features
- [`4f80073`](https://github.com/songloft-org/songloft/commit/4f80073d09ee93e0923593b28c2f4179d5dfdcfc) - 规范设置接口 *(commit by [@hanxi](https://github.com/hanxi))*
- [`7ff3c77`](https://github.com/songloft-org/songloft/commit/7ff3c7722befdf407018c4950bc06a6fbac1bbfa) - 支持设置日志等级 *(commit by [@hanxi](https://github.com/hanxi))*
- [`d3b48ac`](https://github.com/songloft-org/songloft/commit/d3b48ac9442d6bf15c09ca42a3d3bdfccf86a419) - 新增歌词调整 songloft-org/songloft[#49](https://github.com/songloft-org/songloft/pull/49) *(commit by [@hanxi](https://github.com/hanxi))*
- [`b1acd65`](https://github.com/songloft-org/songloft/commit/b1acd65496668ea58e6806c303df0693b869ab17) - 继续优化快速切歌卡顿问题 [#49](https://github.com/songloft-org/songloft/pull/49) *(commit by [@hanxi](https://github.com/hanxi))*
- [`1df746a`](https://github.com/songloft-org/songloft/commit/1df746a6ab63fff9e6d4cc6416f1829315071c71) - 网络歌单转本地优化 *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`c0751a2`](https://github.com/songloft-org/songloft/commit/c0751a255811916d0affee9326778ca641341a89) - 解决m3u8格式的电台无法播放问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`aca70a4`](https://github.com/songloft-org/songloft/commit/aca70a459c805b0f14f0633043532f9966362b7c) - access log 改用 slog 输出，解决日志等级控制问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :construction_worker: Build System
- [`24d74cb`](https://github.com/songloft-org/songloft/commit/24d74cb64ce356ebbfbeeae853d6a53776ce9169) - update *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`0a5af24`](https://github.com/songloft-org/songloft/commit/0a5af24aee6f32d81dd02acc302fddc60bbb1e0e) - update CHANGELOG for v2.3.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`4995c2b`](https://github.com/songloft-org/songloft/commit/4995c2b7c529d99661d9386f87c7ce150acb4f78) - release version 2.4.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.3.0] - 2026-06-01
### :sparkles: New Features
- [`dad2d9c`](https://github.com/songloft-org/songloft/commit/dad2d9cbd100d93bfcb8dce17c408480e32ec562) - 电台支持走后端代理 *(commit by [@hanxi](https://github.com/hanxi))*
- [`2d10351`](https://github.com/songloft-org/songloft/commit/2d103513dc6c808e24f300a3c6440935a246ed26) - 快速切多次歌曲时抖动优化 songloft-org/songloft[#79](https://github.com/songloft-org/songloft/pull/79) *(commit by [@hanxi](https://github.com/hanxi))*
- [`8bedce2`](https://github.com/songloft-org/songloft/commit/8bedce2edb92248f046dd479858b9758e7b5e429) - 添加歌曲支持目录筛选和类型筛选 songloft-org/songloft[#57](https://github.com/songloft-org/songloft/pull/57) *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`42dac46`](https://github.com/songloft-org/songloft/commit/42dac4686c1b4fdd09a56db31e4add94b4c4ca76) - 修复m3u8电台无法播放问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`f0df02e`](https://github.com/songloft-org/songloft/commit/f0df02ee2d1fca28b9246cba94eb9075b09170ab) - update CHANGELOG for v2.2.5 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`fac56c3`](https://github.com/songloft-org/songloft/commit/fac56c3abac8593b4a4c6a5046ac2e41ad0f0d41) - release version 2.3.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.5] - 2026-06-01
### :memo: Documentation Changes
- [`bb299cf`](https://github.com/songloft-org/songloft/commit/bb299cfa04725f097121c2e6d0fef366ceb86c05) - update CHANGELOG for v2.2.4 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`64ca12c`](https://github.com/songloft-org/songloft/commit/64ca12c252de1147713f46940db8997889bd4065) - release version 2.2.5 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.4] - 2026-05-31
### :bug: Bug Fixes
- [`cdaae98`](https://github.com/songloft-org/songloft/commit/cdaae98d90591dec626e64f7cc672e68da7873e4) - 修复ffmpeg转码问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`5e4c594`](https://github.com/songloft-org/songloft/commit/5e4c59477743504cbc945978598f6ee28db39724) - update CHANGELOG for v2.2.3 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`3e9483b`](https://github.com/songloft-org/songloft/commit/3e9483bf39e4a3d54c4c388944647f749b6119b8) - release version 2.2.4 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.3] - 2026-05-31
### :bug: Bug Fixes
- [`cebaeb5`](https://github.com/songloft-org/songloft/commit/cebaeb5131ab3ee37a222fe5c8da88173cae0856) - 修复电台无法播放问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`b9564f7`](https://github.com/songloft-org/songloft/commit/b9564f72cb268f902c907454a7022e2c6acfe920) - 修复ffmpeg转码问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`ed70e1e`](https://github.com/songloft-org/songloft/commit/ed70e1e1270d5dca64646c68e4c767b7ffe695ae) - 上传封面问题 [#78](https://github.com/songloft-org/songloft/pull/78) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`18b47a8`](https://github.com/songloft-org/songloft/commit/18b47a83df81538d755a75082ca327570d8e2f48) - update CHANGELOG for v2.2.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`41b6958`](https://github.com/songloft-org/songloft/commit/41b6958d6d86df25e7f2ba2710fb192d2a426b62) - release version 2.2.3 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.2] - 2026-05-31
### :sparkles: New Features
- [`0b3dbb0`](https://github.com/songloft-org/songloft/commit/0b3dbb0776fbbef743fcacbb4cf4b6a1e940adee) - 优化预加载下一首 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`1187bc9`](https://github.com/songloft-org/songloft/commit/1187bc97baa6c10594edd11318ae10ee6d7ee051) - update CHANGELOG for v2.2.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`8e2c293`](https://github.com/songloft-org/songloft/commit/8e2c293ae654ba2c992ba212fb039a46a9dbfc95) - release version 2.2.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.1] - 2026-05-31
### :bug: Bug Fixes
- [`5843b56`](https://github.com/songloft-org/songloft/commit/5843b560c9dccb080b51a6bbe4481fca1b8787ea) - 解决url参数被解析合并的问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`9cecfe9`](https://github.com/songloft-org/songloft/commit/9cecfe958fd62a7117640c6dd20693756f1717f0) - update CHANGELOG for v2.2.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`987bc83`](https://github.com/songloft-org/songloft/commit/987bc83728ee572d7d5c65214b11e75847cb8213) - release version 2.2.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.2.0] - 2026-05-31
### :bug: Bug Fixes
- [`933bc76`](https://github.com/songloft-org/songloft/commit/933bc7603735379eab1dcba042babe5db73ffcf5) - ffmpeg 执行失败返回原文件 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`4e664c8`](https://github.com/songloft-org/songloft/commit/4e664c8e5e34619f6dba66ad1618247a40bd47fd) - update CHANGELOG for v2.1.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`375416e`](https://github.com/songloft-org/songloft/commit/375416edc175756e512865c0f260bc51533e9f8f) - release version 2.2.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.1.2] - 2026-05-31
### :bug: Bug Fixes
- [`97721e0`](https://github.com/songloft-org/songloft/commit/97721e07d7d11d2e5609867f878f6b8e0f4f4690) - 修复 ffmpeg 转码参数问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`c688fa7`](https://github.com/songloft-org/songloft/commit/c688fa7f121bd6ad5960f49c8720a2a0e01c1477) - update CHANGELOG for v2.1.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`c833fce`](https://github.com/songloft-org/songloft/commit/c833fcee3a8d214e2e67292391c8d50df9166ba9) - release version 2.1.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.1.1] - 2026-05-30
### :memo: Documentation Changes
- [`22f154b`](https://github.com/songloft-org/songloft/commit/22f154bb8449642ab11f9ea0b5ebd42858521c4e) - update CHANGELOG for v2.1.0 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`ceaa65c`](https://github.com/songloft-org/songloft/commit/ceaa65caa7299a6c9b28b940dc5eaf28ce44f133) - release version 2.1.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.1.0] - 2026-05-30
### :sparkles: New Features
- [`b2a19ba`](https://github.com/songloft-org/songloft/commit/b2a19bad067d1dd1a2640b791c0a009d78bfdf68) - 支持运行在sub path下面 [#68](https://github.com/songloft-org/songloft/pull/68) *(commit by [@hanxi](https://github.com/hanxi))*
- [`8038e79`](https://github.com/songloft-org/songloft/commit/8038e7920960ccc853a7dea58ac31cd71aaee51e) - 支持自动转音频编码 [#36](https://github.com/songloft-org/songloft/pull/36) *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`74fd9a5`](https://github.com/songloft-org/songloft/commit/74fd9a54f5c662aeeadfd1edc45142d76b378139) - 修复电台播放问题 [#69](https://github.com/songloft-org/songloft/pull/69) *(commit by [@hanxi](https://github.com/hanxi))*
- [`6b2c484`](https://github.com/songloft-org/songloft/commit/6b2c4841ab3b5bdb14b49cc62017236dafe854c3) - 修复不能单独改密码的问题 [#70](https://github.com/songloft-org/songloft/pull/70) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`587abc9`](https://github.com/songloft-org/songloft/commit/587abc915ec0b803bdaae50ada8b0a81a25b0846) - update CHANGELOG for v2.0.2 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`96b2a91`](https://github.com/songloft-org/songloft/commit/96b2a916b00276f56dc44191f02c068255138c85) - release version 2.1.0 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.0.2] - 2026-05-30
### :sparkles: New Features
- [`aa3cf97`](https://github.com/songloft-org/songloft/commit/aa3cf978a0d0cff3a9e32fc9e5892116af8950a9) - 新增备份和还原功能 *(commit by [@hanxi](https://github.com/hanxi))*
- [`0952e6a`](https://github.com/songloft-org/songloft/commit/0952e6a9889df87550475ad69c9cedd6efaa2b9e) - docker 镜像加入 ffmpeg *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`0d86aef`](https://github.com/songloft-org/songloft/commit/0d86aeff478c3aa08f8fc9f017846d8a377333e3) - 修复 icon 问题 *(commit by [@hanxi](https://github.com/hanxi))*
- [`045c177`](https://github.com/songloft-org/songloft/commit/045c177b20e4cac52700512506bf48e5a0b89c76) - 修复版本发布问题 *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`b8e513e`](https://github.com/songloft-org/songloft/commit/b8e513e7a61e73feef0e29ee00e6740f090ad34d) - update CHANGELOG for v2.0.1 *(commit by [@github-actions[bot]](https://github.com/apps/github-actions))*

### :wrench: Chores
- [`b84f7ba`](https://github.com/songloft-org/songloft/commit/b84f7ba9db5b6eea75dc5737a649d6a50fc218b7) - release version 2.0.2 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.0.1] - 2026-05-29
### :sparkles: New Features
- [`0741adc`](https://github.com/songloft-org/songloft/commit/0741adcd0adc261c6997b42eed62d29f8793b628) - 新增前端调试模式打包 *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`731ab3c`](https://github.com/songloft-org/songloft/commit/731ab3c0dc2cd540cdc6270df6587a52bd122183) - release version 2.0.1 *(commit by [@hanxi](https://github.com/hanxi))*


## [v2.0.0-alpha.1] - 2026-05-29
### :sparkles: New Features
- [`58232ea`](https://github.com/songloft-org/songloft/commit/58232ea174ce9fa6e14cdd1c805bf7410220f0c5) - **app**: one-shot mimusic.db -> songloft.db auto migration (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*

### :bug: Bug Fixes
- [`8e15634`](https://github.com/songloft-org/songloft/commit/8e15634d51f023289afc2c758af9617cec17bf0c) - **ci**: prevent changelog SHA lines from executing as shell commands *(commit by [@hanxi](https://github.com/hanxi))*

### :recycle: Refactors
- [`65c53cb`](https://github.com/songloft-org/songloft/commit/65c53cb356c6fa7d251137288c2188e6abd870bd) - rename Go module mimusic -> songloft (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*
- [`fc2d601`](https://github.com/songloft-org/songloft/commit/fc2d601e7768af24c0e1bdd83628cae5d84e246f) - rename JS plugin global ABI mimusic.* -> songloft.* (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*
- [`c9f05cd`](https://github.com/songloft-org/songloft/commit/c9f05cdb9778f07f08b0df85c0c9b8b33a000469) - rename runtime literals MiMusic -> Songloft (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*
- [`c32ab18`](https://github.com/songloft-org/songloft/commit/c32ab180dbfd24046cf32798840e06dcc31a35d8) - retire jsplugins aggregator, plugins self-host releases *(commit by [@hanxi](https://github.com/hanxi))*

### :construction_worker: Build System
- [`e0ff7b5`](https://github.com/songloft-org/songloft/commit/e0ff7b558e27c56fee4fac03a6bce2c19ed083bb) - complete v2.0 rebrand of release workflows (Phase 3h follow-up) *(commit by [@hanxi](https://github.com/hanxi))*

### :memo: Documentation Changes
- [`1b766d2`](https://github.com/songloft-org/songloft/commit/1b766d2f18b28c3559a5ebaa46bdbe50fa83a138) - add NOTICE and PRIVACY.md for license and privacy compliance *(commit by [@hanxi](https://github.com/hanxi))*
- [`2f5bbf8`](https://github.com/songloft-org/songloft/commit/2f5bbf8f1c4fd2a6e1785e33ec5d49508d6dbec6) - tighten README disclaimers and remove demo site link *(commit by [@hanxi](https://github.com/hanxi))*
- [`85244b1`](https://github.com/songloft-org/songloft/commit/85244b1b7c3814fa5fb3c1afa841f94750649b2d) - announce planned v2.0 rebrand to Songloft *(commit by [@hanxi](https://github.com/hanxi))*
- [`5da054a`](https://github.com/songloft-org/songloft/commit/5da054a440486094375c092cf4d081072ad512fc) - add MIGRATION.md for v2.0 Songloft rebrand *(commit by [@hanxi](https://github.com/hanxi))*
- [`79c6df9`](https://github.com/songloft-org/songloft/commit/79c6df9ded6c6ce17b5231f27d4a82a28a387ac0) - **swagger**: regenerate Swagger after v2.0 rebrand *(commit by [@hanxi](https://github.com/hanxi))*
- [`b3b17d7`](https://github.com/songloft-org/songloft/commit/b3b17d75699e2378300227360fd63dfeb1d53b14) - add V2 release playbook + update-remotes helper *(commit by [@hanxi](https://github.com/hanxi))*
- [`7bc6fdd`](https://github.com/songloft-org/songloft/commit/7bc6fdd5ac37b3ecb0e221e07544fdd1f1fbf2d2) - **playbook**: fix gh CLI command + add transfer helper script *(commit by [@hanxi](https://github.com/hanxi))*
- [`432632d`](https://github.com/songloft-org/songloft/commit/432632d2387719c48c440dad7ef44ae0745e6a68) - complete link migration to songloft-org / songloft.hanxi.cc *(commit by [@hanxi](https://github.com/hanxi))*
- [`392fdc4`](https://github.com/songloft-org/songloft/commit/392fdc407adf447086dfdf877260579252726b6d) - remove 'formerly MiMusic' from AGENTS.md *(commit by [@hanxi](https://github.com/hanxi))*

### :wrench: Chores
- [`07fada7`](https://github.com/songloft-org/songloft/commit/07fada71c7153f481b988f01638bf5beaf6feba3) - remove lxmusic plugins and rename xiaomi plugin (legal cleanup) *(commit by [@hanxi](https://github.com/hanxi))*
- [`a12a765`](https://github.com/songloft-org/songloft/commit/a12a765bb2e7bf713a3dcd988c01861635be0a69) - bump jsplugins submodule to drop xiaomi.json *(commit by [@hanxi](https://github.com/hanxi))*
- [`8027762`](https://github.com/songloft-org/songloft/commit/80277620b88176e082d451727efee52ac415b17c) - bump plugin-toolchain, jsplugin-musicsdk, pkg/tag submodules *(commit by [@hanxi](https://github.com/hanxi))*
- [`6c80981`](https://github.com/songloft-org/songloft/commit/6c80981075acec99ea4fdabd449e634f11b1b54b) - wire up songloft-player submodule rename (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*
- [`abb6ade`](https://github.com/songloft-org/songloft/commit/abb6ade3edfb34a64d98203c9a030b50b4295d2a) - wire up songloft-plugin-miot submodule rename (v2.0) *(commit by [@hanxi](https://github.com/hanxi))*
- [`68a497f`](https://github.com/songloft-org/songloft/commit/68a497f8271a6837d61b995b39841d740e408e43) - gitignore songloft-player-build (Phase 3h follow-up) *(commit by [@hanxi](https://github.com/hanxi))*
- [`95e3c0e`](https://github.com/songloft-org/songloft/commit/95e3c0e3f50864e7156614462bc2162048b92bef) - bump submodule pointers to alpha-published commits *(commit by [@hanxi](https://github.com/hanxi))*


## [1.4.1] - 2026-05-28

### ✨ Features

- `c9f81fe` 默认开启网络歌单自动转本地歌单
- `ea29cdf` 重构歌词接口问题

### 🐛 Bug Fixes

- `0011325` 修复缓存歌曲冲突问题
- `44e5de6` 修复rename文件报错问题

### 🔧 Chores

- `964233c` release version 1.4.1
- `cdf8359` 优化扫描设置开关文案

## [1.4.0] - 2026-05-27

### ✨ Features

- `4483c45` 自动创建歌单功能简化
- `6cfd245` 歌曲下载功能优化
- `f8dcd21` 简化歌曲歌词封面的url逻辑
- `9148dc9` 重构url
- `e8c91b1` 优化url路径
- `7ff56ff` 移除wasm插件模块
- `81c618f` 网络歌曲转本地歌曲支持写入tag

### 🐛 Bug Fixes

- `8de0455` 修复 js fetch 接口问题
- `55e6818` 歌曲去重
- `f612c13` 修复歌单名重复问题
- `e450a56` 修复歌单名重复问题
- `c4951d3` sqlite问题
- `97bc3de` 修复url问题
- `43b2431` 修复插件接口问题

### ♻️ Code Refactoring

- `a37070b` **test**: 删除手写 mock，全切 :memory: 真实 DB
- `c7e2032` **database**: 引入 UnitOfWork，下线 database.Tx/SQLiteTx
- `d58e1d4` **database**: playlist_songs 表切到 PlaylistSongRepository
- `8c23575` **database**: playlists 表切到 PlaylistRepository
- `10337aa` **database**: songs 表切到 SongRepository
- `9d995cc` **database**: js_plugins 仓储改用 sqlc.Queries
- `7094d5f` **database**: configs 表切到 ConfigRepository
- `ea352cd` **database**: tokens 表切到 TokenRepository
- `b004464` **database**: 引入 sqlc + goose + squirrel 基础设施
- `1910fd0` 抽取 InternalURLResolver,让歌词代理 URL 也能带 token 访问

### 📚 Documentation

- `50a67b1` **database**: 新增 DATABASE_MIGRATIONS 操作指南 + 集成 sqlc 命令到 Makefile
- `703d2bc` **agents**: 同步数据库重构后的开发约定

### 🔧 Chores

- `3c07269` release version 1.4.0
- `96aa6c4` bump musicsdk v1.1.0 + lxmusic 用上 LyricFetcher.lyricParams

## [1.3.50] - 2026-05-25

### ✨ Features

- `37ac3b4` 支持网络歌曲转本地歌曲

### 🔧 Chores

- `923b254` release version 1.3.50

## [1.3.49] - 2026-05-24

### 🐛 Bug Fixes

- `4a60ca1` 修复js插件休眠问题

### 🔧 Chores

- `6bcb020` release version 1.3.49

## [1.3.48] - 2026-05-22

### 🐛 Bug Fixes

- `7d8999d` 修复js插件导致宕机问题

### 🔧 Chores

- `16754d4` release version 1.3.48

## [1.3.47] - 2026-05-22

### ✨ Features

- `89eea57` js插件支持手动上传更新

### 🐛 Bug Fixes

- `bac969a` 修复编译警告
- `53e19c0` 修复js异步问题

### 🔧 Chores

- `452aacb` release version 1.3.47

## [1.3.46] - 2026-05-21

### ✨ Features

- `f7b47bc` js插件改成真异步环境
- `65f1164` 优化插件不可用时的提示

### 🔧 Chores

- `3bd3a57` release version 1.3.46

## [1.3.45] - 2026-05-20

### ✨ Features

- `989769c` 自动创建的歌单默认按照数字前缀排序
- `5c47ffc` 新增js虚拟机
- `39dab1b` 新增js api
- `ac27696` 新增 lxmusic 插件

### 🐛 Bug Fixes

- `627f885` 修复关闭进程卡死问题

### 🔧 Chores

- `f9ddbec` release version 1.3.45

## [1.3.43] - 2026-05-16

### 🔧 Chores

- `a9d666a` release version 1.3.43

## [1.3.42] - 2026-05-16

### ✨ Features

- `1349f40` js插件性能优化
- `170a793` js插件支持jsc
- `6058a32` 新增JS插件管理
- `bca3678` js插件开发
- `9a2dc3a` 新增js插件机制
- `1352e6e` 插件休眠更激进

### 🐛 Bug Fixes

- `1528474` 修复js插件相关问题
- `71565f6` 修复js插件问题
- `ea95c15` JS插件问题修复

### ♻️ Code Refactoring

- `c706dbb` **jsplugin**: split playlists permission into read/write

### 🔧 Chores

- `1dafd4a` release version 1.3.42

### 📝 Other Changes

- `e66cf67` log

## [1.3.41] - 2026-05-11

### ✨ Features

- `a04fb2f` 内存优化：空闲插件自动休眠
- `fc60dfd` 内存优化
- `4f77b55` 内存优化

### 🔧 Chores

- `7ca4fad` release version 1.3.41

## [1.3.40] - 2026-05-07

### 🐛 Bug Fixes

- `b055dc0` 修复打包脚本问题

### 🔧 Chores

- `a49aa4c` release version 1.3.40

## [1.3.39] - 2026-05-06

### ✨ Features

- `dd30f31` 歌单排序功能优化，首页歌单数量显示优化，自动生成的歌单名字优化

### 🔧 Chores

- `24c37f9` release version 1.3.39

## [1.3.38] - 2026-05-06

### ✨ Features

- `f886d0c` 新增歌单排序功能
- `a0f0b89` 添加wma格式支持

### 🐛 Bug Fixes

- `d078fa7` 清理失效的本地歌曲
- `3dc4eed` 修复windows网络歌曲无法缓存的问题

### 🔧 Chores

- `85de484` release version 1.3.38

## [1.3.37] - 2026-04-30

### 🐛 Bug Fixes

- `d0b3c2c` 修复vbr播放时长读取错误问题

### 🔧 Chores

- `46592df` release version 1.3.37

## [1.3.35] - 2026-04-29

### ✨ Features

- `a30430c` 优化插件静态资源访问

### 🔧 Chores

- `d72e680` release version 1.3.35

## [1.3.34] - 2026-04-27

### 🐛 Bug Fixes

- `2d0877d` 修复arm/v7系统无法加载插件问题

### 🔧 Chores

- `761c5a1` release version 1.3.34

## [1.3.33] - 2026-04-26

### 🔧 Chores

- `b616fd7` release version 1.3.33

## [1.3.32] - 2026-04-26

### 🐛 Bug Fixes

- `3f5f78d` 修复升级后404问题

### 🔧 Chores

- `04a3278` release version 1.3.32

## [1.3.31] - 2026-04-25

### 🔧 Chores

- `430e88d` release version 1.3.31

## [1.3.30] - 2026-04-25

### 🐛 Bug Fixes

- `b074f73` 兼容 J3455 CPU

### 🔧 Chores

- `973edd1` release version 1.3.30

### 📝 Other Changes

- `8541df4` 插件加载失败添加错误日志

## [1.3.29] - 2026-04-20

### ✨ Features

- `304270f` 插件支持更新
- `fa7e192` 插件支持更新

### 🐛 Bug Fixes

- `60202c9` 修复部分洛雪音源无法使用问题

### 🔧 Chores

- `1a41ee7` release version 1.3.29

## [1.3.28] - 2026-04-20

### ✨ Features

- `27d8ca0` 新增排除目录设置

### 🔧 Chores

- `04bae3b` release version 1.3.28

## [1.3.24] - 2026-04-19

### 🔧 Chores

- `acb5fc2` release version 1.3.24

### 📝 Other Changes

- `6419bcd` 插件超时优化

## [1.3.22] - 2026-04-17

### ✨ Features

- `1110184` 优化启动速度
- `9dc1eda` 删除 entry_path 字段
- `1e880a1` 新增插件重置功能

### 🔧 Chores

- `2fce1be` release version 1.3.22

## [1.3.21] - 2026-04-17

### ✨ Features

- `e7a6779` 优化升级

### 🐛 Bug Fixes

- `cfded04` 修复 FLAC 中的 ID3v2 信息无法解析的问题
- `c488c01` 修复导入相同插件问题

### 🔧 Chores

- `99b5e73` release version 1.3.21

## [1.3.20] - 2026-04-16

### 🔧 Chores

- `0851d64` release version 1.3.20

### 📝 Other Changes

- `1fca16e` 配置国内镜像

## [1.3.18] - 2026-04-15

### ✨ Features

- `dd8887d` 新增批量删除歌单接口
- `5abf830` 缓存功能优化
- `4b74298` 服务端资源缓存优化

### 🐛 Bug Fixes

- `a166e7e` 修复从lite切换到full的问题

### 🔧 Chores

- `cb3b3f7` release version 1.3.18

## [1.3.16] - 2026-04-10

### ✨ Features

- `128aab0` 支持版本回退到底包
- `4c80b7c` 更新后端支持使用代理

### 🐛 Bug Fixes

- `db0e395` 修复升级问题
- `00ff400` 修复升级问题
- `b904424` 修复更新问题
- `3d9ac57` 修复更新问题
- `eb69df6` 修复更新问题
- `7b6b45a` 修复更新问题
- `67e8840` 修复更新问题
- `a09c6a9` 修复端内更新问题

### 🔧 Chores

- `3b6a91d` release version 1.3.16
- `a478a02` release version 1.3.14

### 📝 Other Changes

- `85983bb` 更新问题

## [1.3.13] - 2026-04-09

### ✨ Features

- `4e3ec57` 新增发布内容
- `c912ace` 支持断点续传
- `03a67b4` 新增异步下载接口
- `a381643` 写入 server_platform 到数据库
- `720a06e` 新增执行命令协议
- `d124fd9` 优化无参数启动方式

### 🐛 Bug Fixes

- `d060619` 解决文件权限问题
- `326b618` 网络歌曲导入问题修复
- `1fcce62` 修复导入歌曲问题

### 🔧 Chores

- `303407d` release version 1.3.13

### 📝 Other Changes

- `4cfb584` update doc
- `0ea9351` update doc
- `3baf9e9` 歌单排序优化
- `a2f5968` 调试

## [1.3.12] - 2026-04-08

### ✨ Features

- `e605965` 歌词支持URL类型

### 🔧 Chores

- `a199b72` release version 1.3.12

### 📝 Other Changes

- `f98a23b` 歌词优化

## [1.3.10] - 2026-04-06

### 🐛 Bug Fixes

- `3aee951` 修复报错
- `d80dda2` sql error
- `c908f4e` 修复问题

### ♻️ Code Refactoring

- `12e3c76` 优化扫码登录
- `825ceaa` 优化超时
- `77578cc` 优化网络歌曲播放时长

### 🔧 Chores

- `1579a86` release version 1.3.10

### 📝 Other Changes

- `54ff2b8` update http
- `96559c5` update http
- `daa2ca3` 插件时间问题

## [1.3.9] - 2026-04-03

### ✨ Features

- `ae3865f` add song_count

### 🔧 Chores

- `7f805c2` release version 1.3.9

### 📝 Other Changes

- `a6a551b` 启动优化
- `4c2e0f1` build

## [1.3.8] - 2026-04-03

### ✨ Features

- `cb8a958` 新增并行执行js

### 🔧 Chores

- `bd6a323` release version 1.3.8

### 📝 Other Changes

- `9f727f1` 歌曲缓存目录优化

## [1.3.7] - 2026-04-02

### 🔧 Chores

- `667be2b` release version 1.3.7

### 📝 Other Changes

- `569fc83` 歌词

## [1.3.6] - 2026-04-02

### ♻️ Code Refactoring

- `4e25b28` 优化播放体验

### 🔧 Chores

- `8dc2774` release version 1.3.6

## [1.3.5] - 2026-04-02

### 🔧 Chores

- `91dd27e` release version 1.3.5

## [1.3.4] - 2026-04-01

### ✨ Features

- `5dbf196` 支持上传封面
- `54dcc44` 支持上传封面

### 🐛 Bug Fixes

- `5d253e9` 扫描歌曲宕机问题

### 🔧 Chores

- `6bc0414` release version 1.3.4

## [1.3.3] - 2026-03-31

### ✨ Features

- `8ce8662` 尝试修复lx运行问题

### 🔧 Chores

- `cffd9b5` release version 1.3.3

### 📝 Other Changes

- `ed877c3` delete web
- `e38af01` delete web

## [1.3.2] - 2026-03-30

### ✨ Features

- `663576d` 添加网络歌曲电台接口改为批量

### 🔧 Chores

- `0b7f3a2` release version 1.3.2

### 📝 Other Changes

- `738896f` Update todo list with song-related tasks

## [1.3.1] - 2026-03-30

### 🔧 Chores

- `5c358c7` release version 1.3.1

## [1.3.0] - 2026-03-30

### 🔧 Chores

- `48368d9` release version 1.3.0

## [1.2.8] - 2026-03-30

### ✨ Features

- `e195111` 网络歌曲支持导入图片
- `de1c838` 重构jsruntime
- `b78130f` use ccgo quickjs

### ♻️ Code Refactoring

- `9ceaecc` 优化

### 🔧 Chores

- `9aec414` release version 1.2.8

### 📝 Other Changes

- `fb424e2` 提交wiki
- `3972cad` 接入cqjs
- `cffb54e` 插件健康检测

## [1.2.7] - 2026-03-26

### ✨ Features

- `abff90a` 添加歌曲批量删除 API (POST /songs/batch-delete)

### 🔧 Chores

- `902d4fe` release version 1.2.7

### 📝 Other Changes

- `33d9f2d` update doc
- `9664b49` update doc

## [1.2.6] - 2026-03-25

### 🔧 Chores

- `73a0403` release version 1.2.6

## [1.2.5] - 2026-03-25

### ✨ Features

- `0c438fb` add frontend
- `490db3c` add mobile

### ♻️ Code Refactoring

- `d61dfae` 优化导入速度
- `b1ff8a9` 优化界面

### 🔧 Chores

- `96b42c5` release version 1.2.5
- `7b00668` convert frontend from directory to submodule

### 📝 Other Changes

- `c9d741d` 版本发布脚本
- `c1ce566` 版本发布脚本
- `6025f4b` 新版本
- `54ad790` 新版本
- `51d43a4` 新版本
- `ca53020` 新版本
- `90ed646` 新版本
- `00ff846` 新版本
- `568c5c0` 新版本
- `fdf177e` update frontend
- `029c4d6` 新版本
- `aa66e98` 新版本
- `0833807` frontend 支持独立部署
- `f61d74e` 更新文档
- `72e5c97` 修改名字
- `1b73996` remove mobile
- `d97a16a` update mobile

## [1.2.4] - 2026-03-19

### 🔧 Chores

- `9c70433` release version 1.2.4

## [1.2.3] - 2026-03-19

### ✨ Features

- `dee25c4` 新增清理歌曲功能

### 🐛 Bug Fixes

- `7ee9c74` build failed
- `d8afe4a` 修复paw
- `be6c2c4` 修复通知栏丢失的问题
- `aa01b1d` 修复pwa更新问题
- `154ad14` 修复通知栏消失的问题
- `4bbf98e` 修复乱码问题

### ♻️ Code Refactoring

- `43ff722` 优化界面
- `586049a` 优化移动端播放器
- `1d4b350` 优化移动端播放器
- `43b20a8` 优化移动端播放器
- `7abfe93` 优化移动端播放器
- `2cf5923` 优化移动端播放器
- `e6d7afe` 优化移动端播放器
- `6ee93ee` 优化播放器界面
- `7836361` 重构错误捕获
- `dda851a` 优化播放列表
- `0f3adee` 优化主页
- `3bd566e` 优化日志
- `e1da51d` 优化插件管理
- `1aa8bfa` 优化插件管理
- `892fb58` 优化插件管理

### 🔧 Chores

- `eec45bf` release version 1.2.3
- `84cefea` release version 1.2.2

### 📝 Other Changes

- `8fbf17f` 尝试修复后台通知栏丢失问题
- `30ab1c0` 尝试修复后台通知栏丢失问题
- `53cd756` 强制更新pwa
- `a572312` 测试 tracely sdk
- `eb48e8b` 测试 tracely sdk
- `5309e1c` 测试 tracely sdk
- `af8bfbe` 测试 tracely sdk
- `9991da4` 接入tracely
- `472c300` 接入tracely
- `4c8719f` 细节优化
- `5485f4e` 标题超长则循环滚动
- `c31da28` 修改菜单按钮颜色
- `3f974c7` 尝试修复通知栏消失问题

## [1.2.1] - 2026-02-26

### 🐛 Bug Fixes

- `f9543db` 解决windows网页打不开问题

### 🔧 Chores

- `b042cd3` release version 1.2.1

## [1.2.0] - 2026-02-26

### 🔧 Chores

- `188b602` release version 1.2.0

## [1.1.0] - 2026-02-25

### ✨ Features

- `391d4dd` 新增接口获取token
- `21aeff9` Add mimusic-plugin-musictag as submodule

### 🐛 Bug Fixes

- `893e880` 解决标题问题
- `2092853` 修复乱码问题
- `af4454f` 解决编码乱码问题
- `4894785` 解决编码乱码问题
- `148d36a` 解决编码乱码问题
- `430bb64` 解决编码乱码问题
- `f8cf809` 解决编码乱码问题

### ♻️ Code Refactoring

- `767c806` 优化歌单体验
- `57020cf` 优化图片
- `87a88b7` 优化图片

### 🔧 Chores

- `f5690fc` release version 1.1.0
- `d3cc78b` release version 1.0.12
- `1b9a285` release version 1.0.11
- `e49b9f3` release version 1.0.10

### 📝 Other Changes

- `4d35862` 处理歌曲封面
- `10739d8` 网络歌曲播放时长
- `bf35fa5` close cgo
- `2234e30` update no cgo sqlite

## [1.0.9] - 2026-02-21

### 🔧 Chores

- `e88df7a` release version 1.0.9

## [1.0.8] - 2026-02-21

### 🔧 Chores

- `415b5ef` release version 1.0.8
- `c507f93` release version 1.0.7

## [1.0.6] - 2026-02-21

### 🔧 Chores

- `823f5db` release version 1.0.6
- `34ca5d4` release version 1.0.5
- `caa1448` release version 1.0.4
- `1321d73` release version 1.0.3
- `ad8d269` release version 1.0.2
- `1892189` release version 1.0.1

### 📝 Other Changes

- `338cd69` upate

## [main] - 2026-02-12

### ✨ Features

- `665486d` Add frontend build job and streamline Docker build
- `d959def` Add frontend build job to GitHub Actions workflow
- `be837cf` add web
- `547c3a8` add web
- `f58408d` add web
- `68a50dc` add web
- `ebd21e1` add web
- `069896f` add web
- `518d905` Add mimusic-plugins as a git submodule
- `db470af` 支持CORS
- `34dd18a` add MIT licence

### 🐛 Bug Fixes

- `145c668` fix 颜色
- `22cd7d4` fix 颜色
- `adb8610` fix 颜色
- `043d0d1` fix 颜色
- `244e365` 歌曲读取
- `6720225` 修复数据显示

### 📝 Other Changes

- `488d33e` Refactor GitHub Actions to use bun setup action
- `47d1700` Refactor Docker workflow to use bun setup action
- `3d13902` Remove unnecessary dependency in build-prod target
- `657e3d3` Simplify Dockerfile by removing go mod commands
- `c39f723` clean code
- `09e2957` 简化登录
- `f8fc2c5` 改名为mimusic
[v2.0.0-alpha.1]: https://github.com/songloft-org/songloft/compare/e0a9fd8a53e21bc17982323664e10f8d9549531a...v2.0.0-alpha.1
[v2.0.1]: https://github.com/songloft-org/songloft/compare/v2.0.0...v2.0.1
[v2.0.2]: https://github.com/songloft-org/songloft/compare/v2.0.1...v2.0.2
[v2.1.0]: https://github.com/songloft-org/songloft/compare/v2.0.2...v2.1.0
[v2.1.1]: https://github.com/songloft-org/songloft/compare/v2.1.0...v2.1.1
[v2.1.2]: https://github.com/songloft-org/songloft/compare/v2.1.1...v2.1.2
[v2.2.0]: https://github.com/songloft-org/songloft/compare/v2.1.2...v2.2.0
[v2.2.1]: https://github.com/songloft-org/songloft/compare/v2.2.0...v2.2.1
[v2.2.2]: https://github.com/songloft-org/songloft/compare/v2.2.1...v2.2.2
[v2.2.3]: https://github.com/songloft-org/songloft/compare/v2.2.2...v2.2.3
[v2.2.4]: https://github.com/songloft-org/songloft/compare/v2.2.3...v2.2.4
[v2.2.5]: https://github.com/songloft-org/songloft/compare/v2.2.4...v2.2.5
[v2.3.0]: https://github.com/songloft-org/songloft/compare/v2.2.5...v2.3.0
[v2.4.0]: https://github.com/songloft-org/songloft/compare/v2.3.0...v2.4.0
[v2.5.0]: https://github.com/songloft-org/songloft/compare/v2.4.0...v2.5.0
[v2.5.1]: https://github.com/songloft-org/songloft/compare/v2.5.0...v2.5.1
[v2.6.0]: https://github.com/songloft-org/songloft/compare/v2.5.1...v2.6.0
[v2.6.2]: https://github.com/songloft-org/songloft/compare/v2.6.1...v2.6.2
[v2.6.3]: https://github.com/songloft-org/songloft/compare/v2.6.2...v2.6.3
[v2.6.4]: https://github.com/songloft-org/songloft/compare/v2.6.3...v2.6.4
[v2.7.0]: https://github.com/songloft-org/songloft/compare/v2.6.4...v2.7.0
[v2.8.0]: https://github.com/songloft-org/songloft/compare/v2.7.0...v2.8.0
[v2.8.1]: https://github.com/songloft-org/songloft/compare/v2.8.0...v2.8.1
[v2.8.2]: https://github.com/songloft-org/songloft/compare/v2.8.1...v2.8.2
[v2.8.3]: https://github.com/songloft-org/songloft/compare/v2.8.2...v2.8.3
[v2.8.4]: https://github.com/songloft-org/songloft/compare/v2.8.3...v2.8.4
[v2.8.5]: https://github.com/songloft-org/songloft/compare/v2.8.4...v2.8.5
[v2.8.6]: https://github.com/songloft-org/songloft/compare/v2.8.5...v2.8.6
[v2.8.7]: https://github.com/songloft-org/songloft/compare/v2.8.6...v2.8.7
[v2.8.8]: https://github.com/songloft-org/songloft/compare/v2.8.7...v2.8.8
[v2.8.9]: https://github.com/songloft-org/songloft/compare/v2.8.8...v2.8.9
[v2.8.10]: https://github.com/songloft-org/songloft/compare/v2.8.9...v2.8.10
[v2.9.0]: https://github.com/songloft-org/songloft/compare/v2.8.10...v2.9.0
[v2.9.1]: https://github.com/songloft-org/songloft/compare/v2.9.0...v2.9.1
[v2.9.2]: https://github.com/songloft-org/songloft/compare/v2.9.1...v2.9.2
[v2.9.3]: https://github.com/songloft-org/songloft/compare/v2.9.2...v2.9.3
[v2.9.4]: https://github.com/songloft-org/songloft/compare/v2.9.3...v2.9.4
[v2.9.5]: https://github.com/songloft-org/songloft/compare/v2.9.4...v2.9.5
[v2.9.6]: https://github.com/songloft-org/songloft/compare/v2.9.5...v2.9.6
[v2.10.0]: https://github.com/songloft-org/songloft/compare/v2.9.6...v2.10.0
[v2.11.0]: https://github.com/songloft-org/songloft/compare/v2.10.0...v2.11.0
