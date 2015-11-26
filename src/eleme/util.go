package eleme

import "math/rand"
import "net/http"
import "time"

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

var RandomDataMaker = randomDataMaker{rand.NewSource(time.Now().UnixNano())}

func genRandomString() string {
	ret := make([]byte, 32)
	RandomDataMaker.Read(ret)
	return string(ret[:32])
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

