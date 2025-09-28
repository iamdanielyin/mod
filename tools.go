package mod

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func SplitAndTrimSpace(s, sep string) []string {
	var result []string
	for _, item := range strings.Split(strings.TrimSpace(s), sep) {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func JSONStringify(v any, format ...bool) string {
	var (
		b   []byte
		err error
	)
	if len(format) > 0 && format[0] {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err == nil {
		return string(b)
	}
	return ""
}

func JSONParse(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

func DecodeURIComponent(str string) string {
	if r, err := url.QueryUnescape(str); err == nil {
		return r
	}
	return str
}

func NewRandomNumber(length int) string {
	if length <= 0 {
		return ""
	}

	// 设置随机数生成器的种子
	rand.Seed(time.Now().UnixNano())

	// 第一位不能为0
	firstDigit := rand.Intn(9) + 1 // 1-9 之间的随机数

	// 生成剩余位数的随机数
	randomNumber := fmt.Sprintf("%d", firstDigit)
	for i := 1; i < length; i++ {
		randomNumber += fmt.Sprintf("%d", rand.Intn(10)) // 0-9 之间的随机数
	}

	return randomNumber
}

func NewUUID(upper bool, hyphen bool) string {
	s := uuid.NewString()
	if upper {
		s = strings.ToUpper(s)
	}
	if !hyphen {
		s = strings.ReplaceAll(s, "-", "")
	}
	return s
}

func NewUUIDToken() string {
	return NewUUID(true, false)
}

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(rand.Int63n(1024))
	if err != nil {
		panic(fmt.Errorf("snowflake init failed: %w", err))
	}
}

func NextSnowflakeID() snowflake.ID {
	return snowflakeNode.Generate()
}

func NextSnowflakeIntID() int64 {
	return NextSnowflakeID().Int64()
}

func NextSnowflakeStringID() string {
	return NextSnowflakeID().String()
}

func GenPassword(inputPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	outputPassword := string(hash)
	return outputPassword, err
}

func CheckPassword(inputPassword, targetPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(targetPassword), []byte(inputPassword))
	return err == nil
}

func MD5Str(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseUint(value string) (uint64, error) {
	return strconv.ParseUint(value, 10, 64)
}

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func parseBool(value string) (bool, error) {
	return strconv.ParseBool(value)
}
