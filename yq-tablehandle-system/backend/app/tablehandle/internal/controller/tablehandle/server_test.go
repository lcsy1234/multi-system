package tablehandle

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"testing"
)

func testServer(t *testing.T) *ghttp.Server {
	t.Helper()
	s := ghttp.GetServer(token())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	Bind(s, t.TempDir())
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}
