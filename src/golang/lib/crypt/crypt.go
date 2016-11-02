package crypt

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash"
	"time"
)

type EncodeCtx struct {
	h      hash.Hash
	cipher string
}

func CreateEncodeCtx(passwd string) *EncodeCtx {
	encodeCtx := &EncodeCtx{}
	encodeCtx.h = md5.New()
	encodeCtx.cipher = passwd
	return encodeCtx
}

func cipherEncode(ctx *EncodeCtx, sourceText string) string {
	ctx.h.Write([]byte(ctx.cipher))
	cipherHash := fmt.Sprintf("%x", ctx.h.Sum(nil))
	ctx.h.Reset()
	inputData := []byte(sourceText)
	loopCount := len(inputData)
	outData := make([]byte, loopCount)
	for i := 0; i < loopCount; i++ {
		outData[i] = inputData[i] ^ cipherHash[i%32]
	}
	return string(outData)
}

func Encode(ctx *EncodeCtx, sourceText string) string {
	ctx.h.Write([]byte(time.Now().Format("2015-12-22 15:04:05")))
	noise := fmt.Sprintf("%x", ctx.h.Sum(nil))
	ctx.h.Reset()
	inputData := []byte(sourceText)
	loopCount := len(inputData)
	outData := make([]byte, loopCount*2)

	for i, j := 0, 0; i < loopCount; i, j = i+1, j+1 {
		outData[j] = noise[i%32]
		j++
		outData[j] = inputData[i] ^ noise[i%32]
	}
	return base64.StdEncoding.EncodeToString([]byte(cipherEncode(ctx, fmt.Sprintf("%s", outData))))
}

func Decode(ctx *EncodeCtx, sourceText string) string {
	buf, err := base64.StdEncoding.DecodeString(sourceText)
	if err != nil {
		fmt.Println("Decode(%q) failed: %v", sourceText, err)
		return ""
	}
	inputData := []byte(cipherEncode(ctx, fmt.Sprintf("%s", buf)))
	loopCount := len(inputData)
	outData := make([]byte, loopCount)
	for i, j := 0, 0; i < loopCount; i, j = i+2, j+1 {
		outData[j] = inputData[i] ^ inputData[i+1]
	}
	return string(outData)
}
