package main_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/sharpner/php-multiprocess-proxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Context("test all basic http request types", func() {
		var (
			rec     *httptest.ResponseRecorder
			handler http.HandlerFunc
		)

		BeforeSuite(func() {
			var err error
			rec = httptest.NewRecorder()
			handler, err = NewPHPHTTPHandlerFunc("test/index.php")
			time.Sleep(200 * time.Millisecond)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterSuite(func() {

		})

		It("Should welcome you with hello world", func() {
			req, err := http.NewRequest("GET", "/hello/fisch", nil)
			req.RequestURI = "/hello/fisch"
			req.RemoteAddr = "localhost"
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(ContainSubstring("Hello fisch"))
		})
	})
})
