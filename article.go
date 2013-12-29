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
	PartNum   int64
	PartTotal int64
	PartSize  int64
	PartBegin int64
	PartEnd   int64
	FileNum   int
	FileTotal int
	FileSize  int64
	FileName  string
}

func NewArticle(p []byte, data *ArticleData, subject string) *Article {

	headers := make(map[string]string)

	headers["From"] = Config.Global.From
	headers["Newsgroups"] = Config.Global.DefaultGroup
	// art.headers['Message-ID'] = '<%.5f.%d@%s>' % (time.time(), partnum, self.conf['server']['hostname'])
	headers["X-Newsposter"] = "gopoststuff alpha - https://github.com/madcowfred/gopoststuff"

	// Build subject
	// spec: c1 [fnum/ftotal] - "filename" yEnc (pnum/ptotal)
	headers["Subject"] = fmt.Sprintf("%s [%d/%d] - \"%s\" yEnc (%d/%d)", subject, data.FileNum, data.FileTotal, data.FileName, data.PartNum, data.PartTotal)

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

	return &Article{Headers: headers, Body: buf.Bytes()}
}
