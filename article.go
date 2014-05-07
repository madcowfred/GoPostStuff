package main

import (
	"bytes"
	"fmt"
	"github.com/madcowfred/yencode"
	"hash/crc32"
	"time"
)

type Article struct {
	Body []byte
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
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("From: %s\r\n", Config.Global.From))

	if (len(*groupFlag) > 0) {
		buf.WriteString(fmt.Sprintf("Newsgroups: %s\r\n", *groupFlag))
	} else {
		buf.WriteString(fmt.Sprintf("Newsgroups: %s\r\n", Config.Global.DefaultGroup))
	}
	buf.WriteString(fmt.Sprintf("Message-ID: <%d$gps@gopoststuff>\r\n", time.Now().UnixNano()))
	// art.headers['Message-ID'] = '<%.5f.%d@%s>' % (time.time(), partnum, self.conf['server']['hostname'])
	//headers["X-Newsposter"] = "gopoststuff alpha - https://github.com/madcowfred/gopoststuff"
	buf.WriteString(fmt.Sprintf("X-Newsposter: gopoststuff %s - https://github.com/madcowfred/gopoststuff\r\n", GPS_VERSION))

	// Build subject
	// spec: c1 [fnum/ftotal] - "filename" yEnc (pnum/ptotal)
	var subj string
	if (len(Config.Global.SubjectPrefix) > 0) {
		subj = fmt.Sprintf("%s %s", Config.Global.SubjectPrefix, subject)
	} else {
		subj = subject
	}
	buf.WriteString(fmt.Sprintf("Subject: %s [%d/%d] - \"%s\" yEnc (%d/%d)\r\n\r\n", subj, data.FileNum, data.FileTotal, data.FileName, data.PartNum, data.PartTotal))

	// yEnc begin line
	buf.WriteString(fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", data.PartNum, data.PartTotal, data.FileSize, data.FileName))
	// yEnc part line
	buf.WriteString(fmt.Sprintf("=ypart begin=%d end=%d\r\n", data.PartBegin+1, data.PartEnd))

	//log.Debug("%+v", buf)
	// Encoded data
	yencode.Encode(p, buf)
	// yEnc end line
	h := crc32.NewIEEE()
	h.Write(p)
	buf.WriteString(fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", data.PartSize, data.PartNum, h.Sum32()))

	return &Article{Body: buf.Bytes()}
}
