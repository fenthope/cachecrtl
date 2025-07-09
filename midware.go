package cachecrtl

import (
	"net/http"

	"github.com/infinite-iroha/touka"
)

// CacheControl 中间件的配置选项.
type CacheOptions struct {
	// Rules 是一个路径前缀到 http.Header 的映射.
	Rules map[string]http.Header
	// NoMatchHeaders 是当没有路径匹配任何规则时应用的默认头部.
	// 如果为 nil, 则不应用任何默认头部.
	NoMatchHeaders http.Header
}

type cacheControlMiddleware struct {
	tree           *node // 用于高效匹配的树
	noMatchHeaders http.Header
	enabled        bool
}

// NewCacheControl 创建并初始化一个新的缓存控制中间件.
func NewCacheControl(opts CacheOptions) touka.HandlerFunc {

	m := &cacheControlMiddleware{
		noMatchHeaders: opts.NoMatchHeaders,
	}

	// 优化: 如果没有定义任何规则, 则只应用默认头部(如果有).
	if len(opts.Rules) > 0 {
		m.tree = &node{}
		for path, headers := range opts.Rules {
			// 将 http.Header 作为值存入树中.
			m.tree.AddRule(path, headers)
		}
	}

	return m.handle
}

// handle 是中间件的核心逻辑.
func (m *cacheControlMiddleware) handle(c *touka.Context) {
	var headersToApply http.Header

	// 如果有树(即有规则), 则进行查找.
	if m.tree != nil {
		if value := m.tree.GetValue(c.Request.URL.Path); value != nil {
			// 类型断言, 将 any 转回 http.Header.
			if h, ok := value.(http.Header); ok {
				headersToApply = h
			}
		}
	}

	// 如果没有从树中找到匹配的规则, 使用默认规则.
	if headersToApply == nil {
		headersToApply = m.noMatchHeaders
	}

	// 应用最终确定的头部.
	// 直接遍历 http.Header, 效率更高, 且能正确处理多值头部.
	header := c.Writer.Header()
	for key, values := range headersToApply {
		header.Del(key)
		for _, value := range values {
			header.Add(key, value)
		}
	}

	c.Next()
}
