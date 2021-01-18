package downloader

import (
	"fmt"
	"golang.org/x/net/context"
	"time"
)

type DownloadRateLimiter struct {
	r, n int
}

func NewLimiter(r int) *DownloadRateLimiter {
	return &DownloadRateLimiter{r: r}
}

func (c *DownloadRateLimiter) WaitN(ctx context.Context, n int) (err error) {
	fmt.Println(n)
	c.n += n
	time.Sleep(
		time.Duration(1.00 / float64(c.r) * float64(n) * float64(time.Second)))
	return
}