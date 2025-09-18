package httputils

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type Response struct {
	Code    int         `json:"code"`    // 状态码（默认和 HTTP Status 一致，可拓展为业务码）
	Result  interface{} `json:"result"`  // 数据
	Message string      `json:"message"` // 信息
}

// NewResponse 构造基础响应
func NewResponse() *Response {
	return &Response{}
}

// --- 链式方法 ---
func (r *Response) WithCode(code int) *Response {
	r.Code = code
	return r
}

func (r *Response) WithMessage(msg interface{}) *Response {
	switch v := msg.(type) {
	case error:
		r.Message = v.Error()
	case string:
		r.Message = v
	}
	return r
}

func (r *Response) WithResult(result interface{}) *Response {
	r.Result = result
	return r
}

// --- 快捷响应函数 ---

// Success 返回成功响应
func Success(ctx *app.RequestContext, result interface{}) {
	resp := NewResponse().
		WithCode(http.StatusOK).
		WithMessage("success").
		WithResult(result)

	ctx.JSON(http.StatusOK, resp)
}

// Failed 返回失败响应（默认 400，可自定义）
func Failed(ctx *app.RequestContext, err error, code ...int) {
	status := http.StatusBadRequest
	if len(code) > 0 {
		status = code[0]
	}

	resp := NewResponse().
		WithCode(status).
		WithMessage(err)

	ctx.JSON(status, resp)
}

// AbortFailed 终止请求并返回错误
func AbortFailed(ctx *app.RequestContext, err error, code ...int) {
	Failed(ctx, err, code...)
	ctx.Abort()
}
