package main_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/sharpner/php-multiprocess-proxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getRunningProcessCount() int {
	cmd := "ps ax | grep test/index | grep -v grep | wc -l"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return -1
	}

	countString := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(countString)
	if err != nil {
		return -1
	}

	return count
}

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
			Expect(err).ToNot(HaveOccurred())
		})

		AfterSuite(func() {
			log.Println("Running after Suite")
			pg.Clear()
		})

		It("Should welcome you with hello world", func() {
			requestURI := "/hello/fisch"
			Expect(getRunningProcessCount()).To(Equal(7))
			req, err := http.NewRequest("GET", requestURI, nil)
			req.RequestURI = requestURI
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(ContainSubstring("Hello fisch"))
			Expect(getRunningProcessCount()).To(Equal(7))
		})

		It("Should return 301 but has no location header", func() {
			requestURI := "/301invalid"
			req, err := http.NewRequest("GET", requestURI, nil)
			req.RequestURI = requestURI
			Expect(err).ToNot(HaveOccurred())
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusBadGateway))
			Expect(rec.Body.Bytes()).To(BeNil())
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
