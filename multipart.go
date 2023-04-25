// Handle multipart messages.

package eml

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"os"
	"regexp"
	"strings"
)

type Part struct {
	Type    string
	Charset string
	Data    []byte
	Headers map[string][]string
}

var (
	charsetRx = regexp.MustCompile(`(?is)charset=(.*)`)
)

// Parse the body of a message, using the given content-type. If the content
// type is multipart, the parts slice will contain an entry for each part
// present; otherwise, it will contain a single entry, with the entire (raw)
// message contents.
func parseBody(body []byte, header textproto.MIMEHeader) (parts []Part, err error) {
	contentType := header.Get("Content-Type")
	_, ps, err := mime.ParseMediaType(contentType)
	if err != nil {
		return
	}

	// if mt != "multipart/alternative" {
	// 	parts = append(parts, Part{ct, body, nil})
	// 	return
	// }

	boundary, ok := ps["boundary"]
	if !ok {
		part := decodeBodyPart(body, header)
		parts = append(parts, part)
		return
	}
	r := multipart.NewReader(bytes.NewReader(body), boundary)
	p, err := r.NextPart()
	for err == nil {
		data, _ := io.ReadAll(p) // ignore error
		var subparts []Part
		subparts, err = parseBody(data, p.Header)
		//if err == nil then body have sub multipart, and append him
		if err == nil {
			parts = append(parts, subparts...)
		} else {
			part := decodeBodyPart(data, p.Header)
			parts = append(parts, part)
		}
		p, err = r.NextPart()
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func decodeBodyPart(data []byte, header textproto.MIMEHeader) Part {
	var err error

	contentTransferEncoding := header.Get("Content-Transfer-Encoding")
	if len(contentTransferEncoding) > 0 {
		wasDecoded := false
		switch strings.ToLower(contentTransferEncoding) {
		case "base64":
			data, err = base64.StdEncoding.DecodeString(string(data))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed decode base64: %s\n", err)
			}
			wasDecoded = true
		case "quoted-printable":
			data, _ = io.ReadAll(quotedprintable.NewReader(bytes.NewReader(data)))
			wasDecoded = true
		}

		if wasDecoded {
			header.Del("Content-Transfer-Encoding")
		}
	}

	contentType := header.Get("Content-Type")
	charset := "UTF-8"
	charsetField := charsetRx.FindStringSubmatch(contentType)
	if len(charsetField) > 1 {
		charset = strings.TrimSpace(charsetField[1])
	}

	part := Part{contentType, charset, data, header}
	return part
}
