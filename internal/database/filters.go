package database

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// ConfigFilter 配置过滤条件
type ConfigFilter struct {
	Keyword string
	Limit   int
	Offset  int
	OrderBy string
	Order   string
}

// SongFilter 歌曲过滤条件
type SongFilter struct {
	Type       string
	Keyword    string
	PathPrefix string
	// 标签分类精确过滤（空字符串表示不过滤该维度）。
	Genre    string
	Artist   string
	Album    string
	Language string
	Style    string
	Year     int // 精确年份，0 表示不过滤
	// DecadeStart 年代起始年（如 1990 表示 1990-1999）；0 表示不过滤。
	DecadeStart int
	// ExcludePlaylistLabels 排除属于「带这些 label 的歌单」的歌曲。
	// 典型用途：排除隐藏歌单（label=hidden）里的歌，让主歌曲列表不显示它们。
	ExcludePlaylistLabels []string
	Limit                 int
	Offset                int
	OrderBy               string
	Order                 string
}

// PlaylistFilter 歌单过滤条件
type PlaylistFilter struct {
	Type          string
	Labels        []string
	ExcludeLabels []string
	Keyword       string
	Limit         int
	Offset        int
	OrderBy       string
	Order         string
}

// PlaylistSongFilter 歌单歌曲过滤/排序条件
type PlaylistSongFilter struct {
	Keyword string
	OrderBy string
	Order   string
	Limit   int
	Offset  int
}

// FacetFilter 标签分类聚合（facet）的过滤/排序/分页条件。
type FacetFilter struct {
	// Keyword 对取值做模糊搜索（空表示不过滤）。
	Keyword string
	// OrderBy 排序维度："count"（按歌曲数，默认）或 "name"（按取值名）。
	OrderBy string
	// Order 排序方向 asc/desc；为空时按 OrderBy 取默认（count→desc，name→asc）。
	Order  string
	Limit  int
	Offset int
}

// TokenFilter Token 过滤条件
type TokenFilter struct {
	TokenType string
	IsActive  *bool
	Keyword   string
	Limit     int
	Offset    int
	OrderBy   string
	Order     string
}

// 排序字段白名单：防止 SQL 注入。
// 调用方传入的 OrderBy 必须在白名单内，否则回退到默认排序。
var (
	songOrderWhitelist = map[string]struct{}{
		"id": {}, "title": {}, "artist": {}, "album": {},
		"duration": {}, "added_at": {}, "updated_at": {},
		"file_modified_at": {}, "year": {}, "genre": {},
	}
	playlistOrderWhitelist = map[string]struct{}{
		"id": {}, "name": {}, "position": {},
		"created_at": {}, "updated_at": {},
	}
	configOrderWhitelist = map[string]struct{}{
		"id": {}, "key": {}, "updated_at": {},
	}
	tokenOrderWhitelist = map[string]struct{}{
		"id": {}, "token_type": {}, "expires_at": {}, "created_at": {},
	}
	playlistSongOrderWhitelist = map[string]struct{}{
		"position": {}, "added_at": {}, "title": {},
		"artist": {}, "album": {}, "duration": {}, "updated_at": {},
		"file_modified_at": {},
	}
	playlistSongOrderColumn = map[string]string{
		"position":         "ps.position",
		"added_at":         "ps.added_at",
		"title":            "s.title",
		"artist":           "s.artist",
		"album":            "s.album",
		"duration":         "s.duration",
		"updated_at":       "s.updated_at",
		"file_modified_at": "s.file_modified_at",
	}
	// songFacetColumn 把 facet 维度映射到固定的 SQL 列名/表达式。
	// 仅使用映射值拼接列名（绝不拼用户输入），防 SQL 注入 —— 同 playlistSongOrderColumn 范式。
	songFacetColumn = map[string]string{
		"genre":    "genre",
		"artist":   "artist",
		"album":    "album",
		"language": "language",
		"style":    "style",
		"year":     "year",
		"decade":   "(year / 10) * 10",
	}
)

// facetBaseCond 返回某 facet 维度「取值非空」的基础过滤条件。
// year/decade 用 year>0；文本维度用 <col> != ”。
func facetBaseCond(field, col string) sq.Sqlizer {
	if field == "year" || field == "decade" {
		return sq.Gt{"year": 0}
	}
	return sq.NotEq{col: ""}
}

// applyFacetOrder 对 facet 聚合结果应用排序。
// count→按歌曲数（默认 DESC，附带 value ASC 稳定次序）；name→按取值名（默认 ASC）。
func applyFacetOrder(sb sq.SelectBuilder, f *FacetFilter) sq.SelectBuilder {
	orderBy := "count"
	order := ""
	if f != nil {
		if f.OrderBy != "" {
			orderBy = f.OrderBy
		}
		order = f.Order
	}
	if orderBy == "name" {
		dir := "ASC"
		if strings.EqualFold(order, "DESC") {
			dir = "DESC"
		}
		return sb.OrderBy("value " + dir)
	}
	// 默认按 count
	dir := "DESC"
	if strings.EqualFold(order, "ASC") {
		dir = "ASC"
	}
	return sb.OrderBy("count "+dir, "value ASC")
}

// applyOrder 把 orderBy/order 加到 squirrel SELECT 上。
// orderBy 不在白名单时退化到 defaultOrder（已含 ASC/DESC）。
// tablePrefix 用于带 JOIN 的查询（如 "p."），无前缀传 ""。
func applyOrder(sb sq.SelectBuilder, orderBy, order, defaultOrder string, whitelist map[string]struct{}, tablePrefix string) sq.SelectBuilder {
	if orderBy == "" {
		return sb.OrderBy(defaultOrder)
	}
	if _, ok := whitelist[orderBy]; !ok {
		return sb.OrderBy(defaultOrder)
	}
	dir := "ASC"
	if strings.EqualFold(order, "DESC") {
		dir = "DESC"
	}
	return sb.OrderBy(tablePrefix + orderBy + " " + dir)
}

// applyPlaylistSongOrder 对歌单歌曲查询应用排序。
// 与 applyOrder 不同，使用列名映射表将用户传入的字段名转换为带表别名的列名。
func applyPlaylistSongOrder(sb sq.SelectBuilder, orderBy, order string) sq.SelectBuilder {
	col := "ps.position"
	if orderBy != "" {
		if mapped, ok := playlistSongOrderColumn[orderBy]; ok {
			col = mapped
		}
	}
	dir := "ASC"
	if strings.EqualFold(order, "DESC") {
		dir = "DESC"
	}
	return sb.OrderBy(col + " " + dir)
}

// applyPagination 把 limit/offset 加到 squirrel SELECT 上。limit<=0 视为不分页。
func applyPagination(sb sq.SelectBuilder, limit, offset int) sq.SelectBuilder {
	if limit <= 0 {
		return sb
	}
	sb = sb.Limit(uint64(limit))
	if offset > 0 {
		sb = sb.Offset(uint64(offset))
	}
	return sb
}
