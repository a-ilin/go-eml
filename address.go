// Address parsing

package eml

import (
	"fmt"
	"os"
	"strings"
)

type Address interface {
	String() string
	Name() string
	Email() string
}

type MailboxAddr struct {
	name   string
	local  string
	domain string
}

func (ma MailboxAddr) Name() string {
	if ma.name == "" {
		return fmt.Sprintf("%s@%s", ma.local, ma.domain)
	}
	return ma.name
}

func (ma MailboxAddr) String() string {
	if ma.name == "" {
		if ma.domain == "" {
			return ma.local
		}
		return fmt.Sprintf("%s@%s", ma.local, ma.domain)
	} else if ma.local == "" || ma.domain == "" {
		return ma.name
	}

	return fmt.Sprintf("%s <%s@%s>", ma.name, ma.local, ma.domain)
}

func (ma MailboxAddr) Email() string {
	return fmt.Sprintf("%s@%s", ma.local, ma.domain)
}

type GroupAddr struct {
	name  string
	boxes []MailboxAddr
}

func (ga GroupAddr) Name() string {
	return ga.name
}

func (ga GroupAddr) String() string {
	return ""
}

func (ga GroupAddr) Email() string {
	return ""
}

func ParseAddress(bs []byte) (Address) {
	toks := tokenize(bs)
	return parseAddress(toks)
}

func parseAddress(toks []token) (Address) {
	// If this is a group, it must end in a ";" token.
	ltok := toks[len(toks)-1]
	if len(ltok) == 1 && ltok[0] == ';' {
		ga := GroupAddr{}
		// we split on ':'
		nts, rest := splitOn(toks, []byte{':'})

		for _, nt := range nts {
			ga.name += string(nt) + " "
		}
		ga.name = strings.TrimSpace(ga.name)
		ga.boxes = []MailboxAddr{}

		last := 0
		something := false
		for i, t := range rest {
			if len(t) == 1 && (t[0] == ',' || t[0] == ';') && something {
				ma := parseMailboxAddr(rest[last:i])
				ga.boxes = append(ga.boxes, ma)
				last = i + 1
			}
			something = true
		}
		return ga
	}
	return parseMailboxAddr(toks)
}

func splitOn(ts []token, s token) ([]token, []token) {
	for i, t := range ts {
		if string(t) == string(s) {
			return ts[:i], ts[i+1:]
		}
	}

	fmt.Fprintf(os.Stderr, "Split token not found '%s': %v\n", s, ts)
	return ts, []token{}
}

func parseMailboxAddr(ts []token) (ma MailboxAddr) {
	// We're either name-addr or an addr-spec. If we end in ">", then all
	// characters up to "<" constitute the name. Otherwise, there is no
	// name.
	ma = MailboxAddr{}
	ltok := ts[len(ts)-1]
	if len(ltok) == 1 && ltok[0] == '>' {
		var nts, ats []token
		nts, ats = splitOn(ts, []byte{'<'})

		for _, nt := range nts {
			ma.name += string(nt) + " "
		}
		ma.name = strings.TrimSpace(ma.name)
		ma.name = strings.TrimPrefix(ma.name, `"`)
		ma.name = strings.TrimSuffix(ma.name, `"`)

		if len(ats) > 0 {
			ma.local, ma.domain = parseSimpleAddr(ats[:len(ats)-1])
		}

		return
	}
	ma.local, ma.domain = parseSimpleAddr(ts)
	return
}

func parseSimpleAddr(ts []token) (l, d string) {
	// The second token must be '@' - all further tokens are stuck in the domain.
	if len(ts) > 0 {
		l = string(ts[0])
	}

	if len(ts) > 1 {
		if !(len(ts[1]) == 1 && ts[1][0] == '@') {
			for _, lp := range ts[1:] {
				l += " " + string(lp)
			}
			l = strings.TrimSpace(l)
		} else if len(ts) > 2 {
			for _, dp := range ts[2:] {
				d += string(dp) + " "
			}
			d = strings.TrimSpace(d)
		}
	}
	return
}
