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
		n, err := r.Reader.Read(b[len(b):cap(b)])
		r.Cur = int64(len(b) + n)
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			if r.Total < r.Cur {
				r.Total = r.Cur
			}
			return b, err
		}
	}
}

func (r *VisualReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.Cur += int64(n)
	return n, err
}
