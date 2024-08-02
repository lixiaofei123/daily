package utils

import (
	"crypto/md5"
	"fmt"
	"io"
)

func Md5WithSalt(str string, salt string) string {
	h := md5.New()
	io.WriteString(h, str)
	io.WriteString(h, fmt.Sprintf("(-->%s<--)", salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Md5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Md5Data(data []byte) string {
	h := md5.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
