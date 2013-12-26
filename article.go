package main

import (
	"bytes"
	"fmt"
	"github.com/madcowfred/yencode"
	"hash/crc32"
)

type Article struct {
	Headers map[string]string
	Body    []byte
}

type ArticleData struct {
	PartNum   int
	PartTotal int
	PartSize  uint32
	PartBegin uint32
	PartEnd   uint32
	FileSize  uint32
	FileName  string
}

func NewArticle(p []byte, data *ArticleData) *Article {
	a := &Article{}
	a.Headers["From"] = Config.Global.From
	a.Headers["Newsgroups"] = Config.Global.DefaultGroup
	// FIXME
	a.Headers["Subject"] = "aaa"
	// art.headers['Message-ID'] = '<%.5f.%d@%s>' % (time.time(), partnum, self.conf['server']['hostname'])
	a.Headers["X-Newsposter"] = "gopoststuff alpha - https://github.com/madcowfred/gopoststuff"

	buf := new(bytes.Buffer)
	// yEnc begin line
	buf.WriteString(fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", data.PartNum, data.PartTotal, data.FileSize, data.FileName))
	// yEnc part line
	buf.WriteString(fmt.Sprintf("=ypart begin=%d end=%d", data.PartBegin, data.PartEnd))
	// Encoded data
	yencode.Encode(p, buf)
	// yEnc end line
	h := crc32.NewIEEE()
	h.Write(p)
	buf.WriteString(fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", data.PartSize, data.PartNum, h.Sum32()))

	a.Body = buf.Bytes()

	return a
}
