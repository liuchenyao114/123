package common

import (
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

type verificationValue struct {
	code string
	time time.Time
}

const (
	EmailVerificationPurpose = "v"
	PasswordResetPurpose     = "r"
)

var verificationMutex sync.Mutex
var verificationMap map[string]verificationValue
var verificationMapMaxSize = 10
var VerificationValidMinutes = 10

func GenerateVerificationCode(length int) string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	if length == 0 {
		return code
	}
	return code[:length]
}

func RegisterVerificationCodeWithKey(key string, code string, purpose string) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	// 先无条件清理过期项，避免短时间内大量未过期条目让 map 无限增长
	removeExpiredPairs()
	if len(verificationMap) >= verificationMapMaxSize {
		// 达到硬上限时，删除最早的一条再插入
		var oldestKey string
		var oldestTime time.Time
		for k, v := range verificationMap {
			if oldestKey == "" || v.time.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.time
			}
		}
		if oldestKey != "" {
			delete(verificationMap, oldestKey)
		}
	}
	verificationMap[purpose+key] = verificationValue{
		code: code,
		time: time.Now(),
	}
}

func VerifyCodeWithKey(key string, code string, purpose string) bool {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	value, okay := verificationMap[purpose+key]
	now := time.Now()
	if !okay || int(now.Sub(value.time).Seconds()) >= VerificationValidMinutes*60 {
		return false
	}
	return code == value.code
}

func DeleteKey(key string, purpose string) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	delete(verificationMap, purpose+key)
}

// no lock inside, so the caller must lock the verificationMap before calling!
func removeExpiredPairs() {
	now := time.Now()
	for key := range verificationMap {
		if int(now.Sub(verificationMap[key].time).Seconds()) >= VerificationValidMinutes*60 {
			delete(verificationMap, key)
		}
	}
}

func init() {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap = make(map[string]verificationValue)
}
