package utils

import (
	"errors"
	"fmt"
)

// =============================================================================
// 预导出错误变量（统一使用，禁止在 service/handler 层直接 errors.New）
// =============================================================================

var (
	// 认证/授权类错误
	ErrUnauthorized      = errors.New("未授权或Token失效")
	ErrForbidden         = errors.New("权限不足")
	ErrSchoolNotBound    = errors.New("学校账号未绑定")
	ErrSchoolLoginExpiry = errors.New("学校账号登录已失效")
	ErrTokenExpired      = errors.New("Token已过期")
	ErrInvalidToken      = errors.New("无效的Token")

	// 业务逻辑类错误
	ErrNotFound           = errors.New("资源不存在")
	ErrOperationForbidden = errors.New("操作被禁止")
	ErrConflict           = errors.New("数据已存在")
	ErrStateNotAllowed    = errors.New("当前状态不允许此操作")

	// 数据访问类错误
	ErrDBError    = errors.New("数据库操作失败")
	ErrParseError = errors.New("数据解析失败")

	// 学校服务器代理类错误
	ErrSchoolNoResponse     = errors.New("学校服务器无响应")
	ErrSchoolError          = errors.New("学校服务器返回错误")
	ErrSchoolLoginFailed    = errors.New("学校账号登录失败")
	ErrSchoolMaintenance    = errors.New("学校服务器维护中")
	ErrSchoolPasswordNotSet = errors.New("学校密码未设置")

	// 参数校验类错误
	ErrInvalidParams = errors.New("参数格式错误")
	ErrMissingParams = errors.New("必填参数缺失")

	// 系统/网络类错误
	ErrInternalError      = errors.New("服务器内部错误")
	ErrServiceUnavailable = errors.New("服务暂不可用")
	ErrTimeout            = errors.New("请求超时")
	ErrNetworkError       = errors.New("网络连接失败")
)

// =============================================================================
// 错误码常量 (7位码: 第1-2位=一级类目, 第3-4位=二级类目, 第5-7位=具体错误类型)
// =============================================================================

const (
	ErrCodeSuccess              = 0     // 成功
	ErrCodeUnauthorized         = 10001 // 未授权/Token失效
	ErrCodeForbidden            = 10002 // 权限不足
	ErrCodeSchoolNotBound       = 10003 // 学校账号未绑定
	ErrCodeSchoolLoginExpiry    = 10004 // 学校账号登录失效
	ErrCodeTokenExpired         = 10005 // Token已过期
	ErrCodeInvalidToken         = 10006 // 无效的Token
	ErrCodeNotFound             = 20001 // 资源不存在
	ErrCodeStateNotAllowed      = 20004 // 状态不允许操作
	ErrCodeDBError              = 30001 // 数据库操作失败
	ErrCodeParseError           = 30002 // 数据解析失败
	ErrCodeSchoolNoResponse     = 40001 // 学校服务器无响应
	ErrCodeSchoolError          = 40002 // 学校服务器返回错误
	ErrCodeSchoolLoginFailed    = 40003 // 学校账号登录失败
	ErrCodeSchoolMaintenance    = 40004 // 学校服务器维护中
	ErrCodeSchoolPasswordNotSet = 40005 // 学校密码未设置
	ErrCodeInvalidParams        = 50001 // 参数格式错误
	ErrCodeMissingParams        = 50002 // 必填参数缺失
	ErrCodeInternalError        = 60001 // 服务器内部错误
	ErrCodeServiceUnavailable   = 60002 // 服务暂不可用
	ErrCodeTimeout              = 60003 // 请求超时
	ErrCodeNetworkError         = 60004 // 网络连接失败
)

// WrapMsg 包装错误消息，用于参数校验等无需保留原始错误链的场景
// 如果传入的 err 不为 nil，则返回 fmt.Errorf 链式包装
// 如果传入的 err 为 nil，则返回裸 errors.New（兼容无原始错误的情况）
func WrapMsg(message string, err error) error {
	if err == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %w", message, err)
}
