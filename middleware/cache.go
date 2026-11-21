package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Cache 按请求路径区分缓存策略。
//
// 原先对所有响应一律设置 max-age=604800，会把 index.html 和 /api/* 也缓存一周。
// 后果：服务端升级后，浏览器仍从缓存里拿旧的 index.html → 继续加载旧版本的前端
// 脚本 → 页脚显示旧版本、系统信息显示新版本，并且每次刷新都弹「新版本可用」。
//
// 现在只有 /static/ 下的 CRA 产物可以长期缓存（文件名带内容 hash，内容变了
// 文件名也会变）；首页、消息渲染页、用户上传的 /public/static/ 文件和所有 API
// 一律不缓存。
func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") {
			c.Header("Cache-Control", "public, max-age=604800, immutable")
		} else {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	}
}
