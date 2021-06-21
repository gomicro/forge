package vars

import (
	"testing"

	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestVars(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Vars", func() {
		g.Describe("Set", func() {
			g.It("should set vars", func() {
				e := Vars{}

				e.Set("foo", "bar")
				e.Set("baz", "biz")

				Expect(map[string]string(e)["foo"]).To(Equal("bar"))
				Expect(map[string]string(e)["baz"]).To(Equal("biz"))
			})
		})

		g.Describe("Processing", func() {
			g.It("should process a command and replace the vars", func() {
				e := Vars{}
				e.Set("ID", "af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42")

				cmd := "curl http://do.the.thing/pie/{{.ID}}"

				out := e.Process(cmd)
				Expect(out).To(ContainSubstring("af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42"))
				Expect(out).NotTo(ContainSubstring("{{.ID}}"))
			})

			g.It("should leave unknown vars untouched", func() {
				e := Vars{}
				e.Set("ID", "af248b4b-a0e1-4e0f-94c6-4f0b4e0c7c42")

				cmd := "curl http://do.the.thing/pie/{{.ID}}/{{.Subid}}"

				out := e.Process(cmd)
				Expect(out).NotTo(ContainSubstring("{{.ID}}"))
				Expect(out).To(ContainSubstring("{{.Subid}}"))
			})
		})
	})
}
