package i18n

import (
	"fmt"
	"sync"
)

// Translator 翻译器
type Translator struct {
	lang string
	data map[string]string
	mu   sync.RWMutex
}

var (
	defaultTranslator *Translator
	languages         = make(map[string]map[string]string)
	languagesMu       sync.RWMutex
)

// SetLanguage 设置语言
func SetLanguage(lang string) {
	languagesMu.RLock()
	data, ok := languages[lang]
	languagesMu.RUnlock()

	if !ok {
		// 如果语言不存在，使用默认语言
		lang = "zh_CN"
		languagesMu.RLock()
		data = languages[lang]
		languagesMu.RUnlock()
	}

	defaultTranslator = &Translator{
		lang: lang,
		data: data,
	}
}

// T 获取翻译
func T(key string, args ...interface{}) string {
	if defaultTranslator == nil {
		return key
	}

	defaultTranslator.mu.RLock()
	format, ok := defaultTranslator.data[key]
	defaultTranslator.mu.RUnlock()

	if !ok {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(format, args...)
	}
	return format
}

// RegisterLanguage 注册语言包
func RegisterLanguage(lang string, data map[string]string) {
	languagesMu.Lock()
	defer languagesMu.Unlock()
	languages[lang] = data
}

// GetLanguage 获取当前语言
func GetLanguage() string {
	if defaultTranslator == nil {
		return "zh_CN"
	}
	return defaultTranslator.lang
}

// GetAvailableLanguages 获取可用语言列表
func GetAvailableLanguages() []string {
	languagesMu.RLock()
	defer languagesMu.RUnlock()

	langs := make([]string, 0, len(languages))
	for lang := range languages {
		langs = append(langs, lang)
	}
	return langs
}

// Init 初始化默认语言
func Init() {
	SetLanguage("zh_CN")
}
