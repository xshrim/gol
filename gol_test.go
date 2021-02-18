package gol

import (
	"testing"
	"time"

	"github.com/xshrim/gol/tk"
)

func TestA(t *testing.T) {
	data := M{
		"status-code":    200,
		"latency-time":   2 * time.Second,
		"client-ip":      `10.0.0.1`,
		"request-method": "\\GET",
		"request-uri":    "'www.baidu.com'",
	}

	Prtln(string(map2json(nil, data)))

	ctx := Level(DEBUG).AddFlag(Ljson).With(nil).Field(map[string]interface{}{"traceid": 12345, "userid": 1})
	ctx.Infof("%s is requested by user %s", "/", "tom")
	Prtln(string(ctx.GetField()))

	Prtln(tk.Jsonify(data))
}
