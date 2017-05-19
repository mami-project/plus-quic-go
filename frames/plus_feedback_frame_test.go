package frames

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PLUSFeedbackFrame", func() {
	Context("when parsing", func() {
		It("accepts sample frame", func() {
			b := bytes.NewReader([]byte{0x08, 0x03, 0x01, 0x02, 0x03})
			frame, err := ParsePLUSFeedbackFrame(b)
			Expect(err).ToNot(HaveOccurred())
			Expect(frame.data).To(Equal([]byte{0x01, 0x02, 0x03}))
		})

		It("errors on EOFs", func() {
			data := []byte{0x05, 0x03, 0x01, 0x02}
			_, err := ParsePLUSFeedbackFrame(bytes.NewReader(data))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when writing", func() {
		It("writes a sample frame", func() {
			b := &bytes.Buffer{}
			frame := PLUSFeedbackFrame{Data: []byte{0x00, 0x99, 0x88, 0x77}}
			frame.Write(b, 0)
			Expect(b.Bytes()).To(Equal([]byte{0x08, 0x04, 0x00, 0x99, 0x88, 0x77}))
		})
	})
})
