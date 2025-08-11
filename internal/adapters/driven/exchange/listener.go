package exchange

import (
	"bufio"
	"context"
	"net"
	"time"
)

func Listen(ctx context.Context, addr string, out chan<- []byte) {
	for {
		if ctx.Err() != nil {
			return
		}
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}
		sc := bufio.NewScanner(conn)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			b := sc.Bytes()
			cp := make([]byte, len(b))
			copy(cp, b)
			select {
			case <-ctx.Done():
				_ = conn.Close()
				return
			case out <- cp:
			}
		}
		_ = conn.Close()
		// упадём в начало цикла и переподключимся
	}
}
