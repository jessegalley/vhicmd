package httpclient

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/viper"
)

type countingReader struct {
	r        *bufio.Reader
	uploaded *atomic.Int64
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	if n > 0 {
		cr.uploaded.Add(int64(n))
	}
	return n, err
}

func UploadBigFile(url, token string, data io.Reader) (*http.Response, error) {
	var size int64
	if f, ok := data.(*os.File); ok {
		info, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to get file size: %v", err)
		}
		size = info.Size()
	}

	if size == 0 {
		return nil, fmt.Errorf("refusing to upload empty file (size=0)")
	}

	uploadedBytes := atomic.Int64{}
	cr := &countingReader{
		r:        bufio.NewReaderSize(data, 16*1024*1024), // 16MB buffer
		uploaded: &uploadedBytes,
	}

	req, err := http.NewRequest("PUT", url, cr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	req.Header.Del("Expect")

	if viper.GetBool("debug") {
		fmt.Printf("Starting upload (size: %d bytes)\n", size)
	}

	transport := &http.Transport{
		DisableCompression:    false,
		ForceAttemptHTTP2:     false,
		ExpectContinueTimeout: 2 * time.Second,
		// TCP optimizations
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Control: func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, 1024*1024)
					syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
				})
			},
		}).DialContext,
		WriteBufferSize: 64 * 1024 * 1024,
		ReadBufferSize:  64 * 1024 * 1024,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	startTime := time.Now()
	stopProgress := make(chan struct{})
	go trackProgress(&uploadedBytes, size, startTime, stopProgress)

	resp, err := client.Do(req)
	close(stopProgress)

	if err != nil {
		return nil, fmt.Errorf("upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func trackProgress(uploadedBytes *atomic.Int64, size int64, startTime time.Time, stopChan chan struct{}) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer tw.Flush()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastBytes int64
	lastTime := startTime
	var emaSpeed float64
	alpha := 0.5

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			current := uploadedBytes.Load()
			diff := current - lastBytes
			elapsed := now.Sub(lastTime).Seconds()
			totalElapsed := now.Sub(startTime)

			lastBytes = current
			lastTime = now

			// Calculate instantaneous speed and smooth it
			currentSpeed := float64(diff) / elapsed
			if emaSpeed == 0 {
				emaSpeed = currentSpeed
			} else {
				emaSpeed = alpha*currentSpeed + (1-alpha)*emaSpeed
			}

			// Calculate ETA using smoothed speed
			eta := "N/A"
			if emaSpeed > 0 {
				bytesLeft := float64(size - current)
				secsLeft := bytesLeft / emaSpeed
				eta = formatDuration(time.Duration(secsLeft) * time.Second)
			}

			// Print progress
			fmt.Fprintf(tw, "\r\033[KElapsed:\t%s\tUploaded:\t%.1f%%\tSpeed:\t%s/s\tETA:\t%s\t(%d/%d MB)",
				formatDuration(totalElapsed),
				float64(current)*100/float64(size),
				speedToString(emaSpeed),
				eta,
				current/1024/1024,
				size/1024/1024,
			)
			tw.Flush()

			if current >= size {
				fmt.Fprintln(tw)
				tw.Flush()
				return
			}

		case <-stopChan:
			return
		}
	}
}

// speedToString converts bytes/sec into a human-readable (KB/s or MB/s) string
func speedToString(bps float64) string {
	switch {
	case bps < 1024:
		return fmt.Sprintf("%.0f B", bps)
	case bps < 1024*1024:
		return fmt.Sprintf("%.1f KB", bps/1024)
	default:
		return fmt.Sprintf("%.1f MB", bps/(1024*1024))
	}
}

// formatDuration formats a time.Duration as HH:MM:SS
func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
