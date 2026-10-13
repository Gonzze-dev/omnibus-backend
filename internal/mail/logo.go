package mail

import (
	_ "embed"
	"io"
	"strings"

	"gopkg.in/gomail.v2"
)

//go:embed isotype.png
var isotypePNG []byte

// isotypeCID is the Content-ID templates use to reference the isotype (src="cid:isotype.png").
const isotypeCID = "isotype.png"

// embedIsotype attaches the isotype inline when the HTML body references it.
func embedIsotype(msg *gomail.Message, body string) {
	if !strings.Contains(body, "cid:"+isotypeCID) {
		return
	}
	msg.Embed(isotypeCID,
		gomail.SetHeader(map[string][]string{"Content-Type": {"image/png"}}),
		gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(isotypePNG)
			return err
		}),
	)
}
