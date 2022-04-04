package fileutil

import "io"

// VisualReader 将普通包装为可查看进度的Reader
type VisualReader struct {
	io.Reader
	Total int64
	Cur   int64
}

func (r *VisualReader) ReadAll() ([]byte, error) {
	if r.Total <= 0 {
		r.Total = 512
	}
	b := make([]byte, 0, r.Total)

	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		r.Cur = int64(len(b) + n)
		if r.Cur > r.Total {
			r.Total += 512
		}
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			r.Total = int64(len(b))
			return b, err
		}
	}
}

func (r *VisualReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Cur += int64(n)
	return
}
