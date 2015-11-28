package eleme

import "math/rand"
import "net/http"
import "time"
import "crypto/md5"
import "encoding/hex"

type randomDataMaker struct {
    src rand.Source
}

const printable = "abcdefghijABCDEFGHIJKLMNOPQRSTUV"

func (r *randomDataMaker) Read(p []byte) (n int, err error) {
    todo := len(p)
    offset := 0
    for {
        val := int64(r.src.Int63())
        for i := 0; i < 8; i++ {
            p[offset] = printable[val & 0x1f]
            todo--
            if todo == 0 {
                return len(p), nil
            }
            offset++
            val >>= 8
        }
    }

    panic("unreachable")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func (r *randomDataMaker) RandStringBytesMaskImprSrc(n int) string {
    b := make([]byte, n)
    // A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
    for i, cache, remain := n-1, r.src.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = r.src.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }

    return string(b)
}

var RandomDataMaker = randomDataMaker{rand.NewSource(time.Now().UnixNano())}

func genRandomString() string {
	/*
	ret := make([]byte, 16)
	RandomDataMaker.Read(ret)
	return string(ret[:16])
	*/
	return RandomDataMaker.RandStringBytesMaskImprSrc(16)
}

func md5_hex(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func auth(w http.ResponseWriter, r *http.Request) string {
	token := r.URL.Query().Get("access_token")
	if token == "" {
		token = r.Header.Get("Access-Token")	
	}
	if token == "" {
		invalidToken(w)
		return ""
	}
	userid := userModel.findUserIdByToken(token)
	if userid == "" { 
		invalidToken(w)
		return ""
	}
	return userid
}

