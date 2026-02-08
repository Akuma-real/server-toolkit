package i18n

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLanguageAndTranslateConcurrently(t *testing.T) {
	Init()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				SetLanguage("zh_CN")
			} else {
				SetLanguage("en_US")
			}
		}(i)

		go func() {
			defer wg.Done()
			_ = T("app_title")
			_ = GetLanguage()
		}()
	}

	wg.Wait()

	title := T("app_title")
	assert.NotEmpty(t, title)
}
