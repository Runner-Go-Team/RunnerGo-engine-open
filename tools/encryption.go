package tools

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

type Encryption interface {
	HashFunc(str string) (data string)
}

type Md5Type struct {
}

func (md Md5Type) HashFunc(str string) string {
	return MD5(str)
}

type Sha256Type struct {
}

func (sh Sha256Type) HashFunc(str string) string {
	return SHA256(str)
}

type Sha512Type struct {
}

func (sh Sha512Type) HashFunc(str string) string {
	return SHA512(str)
}

func GetEncryption(str string) (encryption Encryption) {
	if str == "MD5" || str == "MD5-sess" {
		encryption = Md5Type{}
		return
	}
	if str == "SHA-256" || str == "SHA-256-sess" {
		encryption = Sha256Type{}
		return
	}
	if str == "SHA-512-256" || str == "SHA-512-256-sess" {
		encryption = Sha512Type{}
	}
	return

}

func MD5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA512(str string) string {
	h := sha512.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA224(str string) string {
	h := sha256.New224()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA384(str string) string {
	h := sha512.New384()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
