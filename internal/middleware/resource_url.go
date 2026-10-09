package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"sop/global"
)

// resourcePathPattern 匹配 JSON 字符串值中的相对资源路径（/uploads/...），
// 仅匹配引号包裹、且其后紧跟引号或反斜杠转义（确保是独立字符串值，不做部分匹配）。
// 示例："/uploads/media/a.png"、"\/uploads\/media\/b.png"
var resourcePathPattern = regexp.MustCompile(`"/(uploads/[^"\\]*)"`)

// absoluteURLPattern 匹配请求体中的完整资源 URL（http(s)://host[:port]/uploads/...），
// 用于写接口提交时剥回相对路径，避免绝对地址入库。
var absoluteURLPattern = regexp.MustCompile(`https?://[^"/\s]*/(uploads/[^"\\]*)"`)

// resourceURLBodyWriter 缓冲响应体，等待 handler 执行完后再做统一改写，
// 避免 JSON 分片导致跨 chunk 匹配失败。
type resourceURLBodyWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *resourceURLBodyWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

// Unwrap 支持 http.ResponseController 等穿透到真实 writer（与项目其他中间件约定一致）
func (w *resourceURLBodyWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *resourceURLBodyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *resourceURLBodyWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// buildResourceBase 组装资源完整 URL 前缀：优先 staticPort，未配置时回退业务端口。
// baseUrl 未配置时返回空串，此时中间件保持相对路径不变。
func buildResourceBase() string {
	srv := global.GetConfig().Server
	if srv.BaseURL == "" {
		return ""
	}
	port := srv.StaticPort
	if port <= 0 {
		port = srv.Port
	}
	return srv.BaseURL + ":" + itoa(port)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// rewriteBody 把响应体中的相对资源路径替换为完整 URL。
// 幂等：仅替换仍为相对路径的条目；空 base 时原样返回，避免生成壳 URL。
func rewriteBody(body []byte, base string) []byte {
	if base == "" {
		return body
	}
	return resourcePathPattern.ReplaceAllFunc(body, func(m []byte) []byte {
		// m 形如 "/uploads/xxx.png"（含首尾引号），替换引号内路径。
		// 整体替换：{" + absolute ，保持闭合引号。
		inner := m[1 : len(m)-1]
		out := make([]byte, 0, len(inner)+len(base))
		out = append(out, '"')
		out = append(out, base...)
		out = append(out, inner...)
		out = append(out, '"')
		return out
	})
}

// stripRequestBody 把请求体中已带域名的完整资源 URL 剥回相对路径（/uploads/...），
// 避免写接口把绝对地址原样入库。幂等：不带域名的不动。
func stripRequestBody(body []byte) []byte {
	return absoluteURLPattern.ReplaceAll(body, []byte(`"$1"`))
}

// ResourceURL 改写 API 出参中的资源相对路径为完整 URL（response 方向），
// 并对写接口的请求体做反向剥回（request 方向），防止绝对地址入库污染。
//
// 基础地址/端口直接读取全局配置 server.baseUrl / server.staticPort / server.port，
// 调用方无需传参；挂在单个路由或全局组上均可，重复挂载因幂等保护不会重复拼接。
//
// 说明：
//  1. 只对引号包裹的 "/uploads/..." JSON 字符串值生效，正文文本不受影响；
//  2. request 方向仅对写接口生效取决于挂载位置，挂读接口时请求体为空则自然跳过；
//  3. 响应体在内存中缓冲，完成后一次性写出，避免分片匹配失败；大响应会有内存开销，
//     故不应挂在 /uploads 等大文件传输路由上。
func ResourceURL() gin.HandlerFunc {
	base := buildResourceBase()
	return func(c *gin.Context) {
		// request 方向：读接口请求体通常为空，代价极小；Write 接口在此剥回绝对地址。
		if c.Request.Body != nil && c.Request.Body != http.NoBody {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil && len(body) > 0 {
				if stripped := stripRequestBody(body); len(stripped) > 0 {
					c.Request.Body = io.NopCloser(bytes.NewReader(stripped))
				}
			}
		}

		// response 方向：缓冲改写后写给真实 writer
		writer := &resourceURLBodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		out := rewriteBody(writer.buf.Bytes(), base)
		if len(out) > 0 {
			_, _ = writer.ResponseWriter.Write(out)
		}
	}
}
