package hostname

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	maxHostnameLength = 253
	maxSegmentLength  = 63
)

// ValidateHostname 验证主机名（RFC 1123）
func ValidateHostname(name string) error {
	if name == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// 转小写
	name = strings.ToLower(name)

	// 检查总长度
	if len(name) > maxHostnameLength {
		return fmt.Errorf("hostname too long (max %d characters)", maxHostnameLength)
	}

	// 检查格式：允许 a-z0-9- 和 .
	// 每段不能以 - 开头或结尾
	// 每段长度 <= 63

	// 检查基本格式
	matched, err := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`, name)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid hostname format")
	}

	// 检查每段长度
	segments := strings.Split(name, ".")
	for _, seg := range segments {
		if len(seg) > maxSegmentLength {
			return fmt.Errorf("hostname segment too long (max %d characters)", maxSegmentLength)
		}
	}

	return nil
}

// IsFQDN 判断是否为 FQDN
func IsFQDN(name string) bool {
	return strings.Contains(name, ".")
}

// GetShortHostname 从 FQDN 中提取短主机名
func GetShortHostname(fqdn string) string {
	idx := strings.Index(fqdn, ".")
	if idx == -1 {
		return fqdn
	}
	return fqdn[:idx]
}

// NormalizeHostname 标准化主机名（转小写，去除多余空格）
func NormalizeHostname(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	return name
}

// ValidateFQDN 验证 FQDN
func ValidateFQDN(fqdn string) error {
	// 先验证基本格式
	if err := ValidateHostname(fqdn); err != nil {
		return fmt.Errorf("invalid FQDN: %w", err)
	}

	// FQDN 必须包含点
	if !strings.Contains(fqdn, ".") {
		return fmt.Errorf("FQDN must contain at least one dot")
	}

	return nil
}
