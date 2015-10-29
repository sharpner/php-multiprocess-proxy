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
			pg      ProcessGroup
		)

		BeforeEach(func() {
			rec = httptest.NewRecorder()
		})

		BeforeSuite(func() {
			var err error
			handler, pg, err = NewPHPHTTPHandlerFunc("test/index.php")
			time.Sleep(200 * time.Millisecond)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterSuite(func() {
			pg.Clear()
		})

		It("Should welcome you with hello world", func() {
			requestURI := "/hello/fisch"
			req, err := http.NewRequest("GET", requestURI, nil)
			req.RequestURI = requestURI
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(ContainSubstring("Hello fisch"))
		})

		It("Should return 301", func() {
			requestURI := "/301"
			req, err := http.NewRequest("GET", requestURI, nil)
			req.RequestURI = requestURI
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusMovedPermanently))
			Expect(rec.Body.Bytes()).To(BeNil())
		})
	})
})
