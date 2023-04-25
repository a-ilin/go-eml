package decoder

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"

	"encoding/base64"
	"mime/quotedprintable"

	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
)

var  (
	parseRx = regexp.MustCompile(`^=\?(.*?)\?(.*?)\?(.*)\?=$`)
)

func UTF8(cs string, data []byte) ([]byte, error) {
	if strings.ToUpper(cs) == "UTF-8" {
		return data, nil
	}

	r, err := charset.NewReader(cs, bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(r)

}

func parseField(bstr []byte) ([]byte, error) {
	var err error
	strs := parseRx.FindAllStringSubmatch(string(bstr), -1)

	if len(strs) > 0 && len(strs[0]) == 4 {
		c := strs[0][1]
		e := strs[0][2]
		dstr := strs[0][3]

		bstr, err = Decode(e, []byte(dstr))
		if err != nil {
			return bstr, err
		}

		return UTF8(c, bstr)
	}
	return bstr, err
}

func Parse(bstr []byte) ([]byte, error) {
	var err error
	strs := parseRx.FindAllStringSubmatch(string(bstr), -1)

	if len(strs) > 0 && len(strs[0]) == 4 {
		multistr := []byte{}

		// decode multiline string separately
		sbs := strings.Fields(string(bstr))
		for _, sb := range sbs {
			sbd, err := parseField([]byte(sb))
			if err != nil {
				return bstr, err
			}
			multistr = append(multistr, sbd...)
		}

		return multistr, nil
	}
	return bstr, err
}

func Decode(e string, bstr []byte) ([]byte, error) {
	var err error
	switch strings.ToUpper(e) {
	case "Q":
		bstr, err = ioutil.ReadAll(quotedprintable.NewReader(bytes.NewReader(bstr)))
	case "B":
		bstr, err = base64.StdEncoding.DecodeString(string(bstr))
	default:
		//not set encoding type

	}
	return bstr, err
}
